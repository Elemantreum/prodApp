package user

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"prodApp/internal/platform/apperrors"
	http2 "prodApp/internal/platform/http"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, user *User) error
	FindById(ctx context.Context, id int) (*User, error)
	GetAll(ctx context.Context) ([]User, error)
	FindAdmins(ctx context.Context) ([]User, error)
}
type IdentityRepository interface {
	Create(ctx context.Context, identity *Identity) (*Identity, error)
	//Update(ctx context.Context, identity *Identity) (*Identity, error)
	//Delete(ctx context.Context, identity *Identity) error
	//FindById(ctx context.Context, id int) (*Identity, error)
	FindByProviderAndExternalID(ctx context.Context, provider, externalID string) (*Identity, error)
	FindByUserID(ctx context.Context, userID int) ([]Identity, error)
	Reassign(ctx context.Context, provider, externalID string, newUserID int) error
	GetBatchById(ctx context.Context, ids []int, provider string) ([]Identity, error)
}

type LinkCodeRepository interface {
	Create(ctx context.Context, code *LinkCode) error
	FindAndDelete(ctx context.Context, code string) (*LinkCode, error)
}

// BookingChecker — узкий срез booking.BookingService, нужный только для
// решения "можно ли тихо удалить осиротевший после привязки аккаунт". Не
// импортируем сам пакет booking напрямую — он уже импортирует user (ради
// UserRepository), так что было бы cycle; main.go подсовывает сюда
// *booking.BookingService, который эту сигнатуру уже удовлетворяет.
type BookingChecker interface {
	HasBookings(ctx context.Context, userID int) (bool, error)
}

type UserService struct {
	userRepo       UserRepository
	identityRepo   IdentityRepository
	linkCodeRepo   LinkCodeRepository
	bookingChecker BookingChecker
}

func NewUserService(userRepo UserRepository, identityRepo IdentityRepository, linkCodeRepo LinkCodeRepository, bookingChecker BookingChecker) *UserService {
	return &UserService{userRepo: userRepo, identityRepo: identityRepo, linkCodeRepo: linkCodeRepo, bookingChecker: bookingChecker}
}

const (
	linkCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // без 0/O/1/I — легко перепутать при вводе руками
	linkCodeLength   = 8
	linkCodeTTL      = 5 * time.Minute
)

// GenerateLinkCode выпускает одноразовый код для уже залогиненного юзера —
// его нужно ввести при входе с ДРУГОЙ платформы, чтобы схлопнуть оба
// аккаунта в один (см. ResolveAndLink).
func (s *UserService) GenerateLinkCode(ctx context.Context, userID int) (*LinkCode, error) {
	buf := make([]byte, linkCodeLength)
	for i := range buf {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(linkCodeAlphabet))))
		if err != nil {
			return nil, fmt.Errorf("generate link code: %w", err)
		}
		buf[i] = linkCodeAlphabet[n.Int64()]
	}

	code := &LinkCode{
		Code:      string(buf),
		UserId:    userID,
		ExpiresAt: time.Now().Add(linkCodeTTL),
	}

	if err := s.linkCodeRepo.Create(ctx, code); err != nil {
		return nil, fmt.Errorf("save link code: %w", err)
	}

	return code, nil
}

// ResolveAndLink погашает код (одноразово — FindAndDelete забирает его
// атомарно) и привязывает platform-идентичность к аккаунту, на который код
// был выписан. В отличие от LinkIdentity — не отказывает, если эта
// идентичность уже привязана к ДРУГОМУ (скорее всего только что
// автосозданному) аккаунту, а молча переставляет её: сюда попадают только
// через одновременное подтверждение владения обеими сторонами (валидная
// подпись initData/launch params + код, который знает только владелец
// исходного аккаунта), так что переставлять безопасно.
func (s *UserService) ResolveAndLink(ctx context.Context, code, provider, externalID string) (*User, error) {
	lc, err := s.linkCodeRepo.FindAndDelete(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("invalid link code: %w", err)
	}

	if time.Now().After(lc.ExpiresAt) {
		return nil, &apperrors.ValidationError{Field: "code: истёк, запросите новый"}
	}

	// Если эта идентичность уже была на ком-то ДРУГОМ (обычный случай — юзер
	// открыл вторую платформу до ввода кода, она автосоздала себе отдельный
	// аккаунт) — запомним его, чтобы после переноса проверить, не осиротел ли.
	var previousOwnerID int
	if existing, err := s.identityRepo.FindByProviderAndExternalID(ctx, provider, externalID); err == nil {
		previousOwnerID = existing.UserId
	}

	if err := s.identityRepo.Reassign(ctx, provider, externalID, lc.UserId); err != nil {
		return nil, fmt.Errorf("reassign identity: %w", err)
	}

	if previousOwnerID != 0 && previousOwnerID != lc.UserId {
		s.cleanupOrphan(ctx, previousOwnerID)
	}

	return s.GetUser(ctx, lc.UserId)
}

func (s *UserService) GetBatchTgById(ctx context.Context, id []int, provider string) ([]Identity, error) {
	identity, err := s.identityRepo.GetBatchById(ctx, id, provider)
	if err != nil {
		return nil, fmt.Errorf("get identity by id: %w", err)
	}
	return identity, err
// cleanupOrphan подчищает аккаунт, оставшийся без единой привязки после
// ResolveAndLink. Удаляет ТОЛЬКО если на нём нет ни одной брони — Delete
// каскадно снёс бы их вместе с аккаунтом, а это уже настоящая потеря
// данных, если юзер успел записаться на что-то до того как решил привязаться.
// В этом случае лучше молча оставить дубликат — админ увидит его в списке
// пользователей и разберётся руками. Любая ошибка здесь — best-effort,
// не должна ломать сам ResolveAndLink (identity уже успешно перенесена).
func (s *UserService) cleanupOrphan(ctx context.Context, userID int) {
	remaining, err := s.identityRepo.FindByUserID(ctx, userID)
	if err != nil || len(remaining) > 0 {
		return
	}

	hasBookings, err := s.bookingChecker.HasBookings(ctx, userID)
	if err != nil || hasBookings {
		return
	}

	orphan, err := s.userRepo.FindById(ctx, userID)
	if err != nil {
		return
	}

	_ = s.userRepo.Delete(ctx, orphan)
}

// EnsureProfile находит пользователя, привязанного к данной платформенной identity,
// либо создаёт нового пользователя вместе с первой identity, если такой ещё нет.
func (s *UserService) EnsureProfile(ctx context.Context, provider, externalID, firstName, lastName string) (*User, error) {
	identity, err := s.identityRepo.FindByProviderAndExternalID(ctx, provider, externalID)
	if err == nil {
		return s.userRepo.FindById(ctx, identity.UserId)
	}
	if !errors.Is(err, apperrors.NotFoundError{}) {
		return nil, fmt.Errorf("find identity: %w", err)
	}

	// identity не найдена — создаём нового пользователя + первую привязку
	newUser := &User{
		Username:   fmt.Sprintf("%s:%s", provider, externalID),
		FirstName:  firstName,
		LastName:   lastName,
		IsApproved: false,
		IsAdmin:    false,
		// Даже если платформа уже дала имя (Telegram обычно даёт, VK — нет),
		// заставляем явно пройти шаг регистрации — см. UpdateProfile ниже.
		IsRegistered: false,
	}

	createdUser, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	_, err = s.identityRepo.Create(ctx, &Identity{
		UserId:     createdUser.Id,
		Provider:   provider,
		ExternalId: externalID,
	})
	if err != nil {
		return nil, fmt.Errorf("create identity: %w", err)
	}

	return createdUser, nil
}

// LinkIdentity привязывает НОВЫЙ способ входа к УЖЕ существующему пользователю —
// используется во флоу "привязать Telegram к своему аккаунту VK" через код привязки.
func (s *UserService) LinkIdentity(ctx context.Context, userID int, provider, externalID string) error {
	existing, err := s.identityRepo.FindByProviderAndExternalID(ctx, provider, externalID)
	if err == nil {
		if existing.UserId != userID {
			return &apperrors.ConflictError{
				Resource: "identity",
				Reason:   "already linked to a different account",
			}
		}
		return nil // уже привязано к этому же юзеру — идемпотентно
	}
	if !errors.Is(err, apperrors.NotFoundError{}) {
		return fmt.Errorf("find identity: %w", err)
	}

	_, err = s.identityRepo.Create(ctx, &Identity{
		UserId:     userID,
		Provider:   provider,
		ExternalId: externalID,
	})
	if err != nil {
		return fmt.Errorf("link identity: %w", err)
	}
	return nil
}

func (s *UserService) AuthenticateViaTelegram(ctx context.Context, telegramID string, firstName string, lastName string) (*User, error) {
	identity, err := s.identityRepo.FindByProviderAndExternalID(ctx, "telegram", telegramID)
	if err == nil {
		return s.userRepo.FindById(ctx, identity.UserId)
	}
	if !errors.Is(err, apperrors.NotFoundError{}) {
		return nil, fmt.Errorf("find identity: %w", err)
	}

	user, err := s.userRepo.Create(ctx, &User{FirstName: firstName, LastName: lastName})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	_, err = s.identityRepo.Create(ctx, &Identity{
		UserId:     user.Id,
		Provider:   "telegram",
		ExternalId: telegramID,
	})

	if err != nil {
		return nil, fmt.Errorf("create identity: %w", err)
	}

	return user, nil
}

func (s *UserService) AuthenticateViaVK(ctx context.Context, telegramID string, firstName string, lastName string) (*User, error) {
	identity, err := s.identityRepo.FindByProviderAndExternalID(ctx, "vk", telegramID)
	if err == nil {
		return s.userRepo.FindById(ctx, identity.UserId)
	}
	if !errors.Is(err, apperrors.NotFoundError{}) {
		return nil, fmt.Errorf("find identity: %w", err)
	}

	user, err := s.userRepo.Create(ctx, &User{FirstName: firstName, LastName: lastName})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	_, err = s.identityRepo.Create(ctx, &Identity{
		UserId:     user.Id,
		Provider:   "vk",
		ExternalId: telegramID,
	})

	if err != nil {
		return nil, fmt.Errorf("create identity: %w", err)
	}

	return user, nil
}

func (service *UserService) NewUser(ctx context.Context, firstName string, lastName string) (*User, error) {

	var user *User = &User{FirstName: firstName, LastName: lastName}

	saved, err := service.userRepo.Create(ctx, user)

	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return saved, nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]User, error) {
	users, err := s.userRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *UserService) ListAdmins(ctx context.Context) ([]User, error) {
	users, err := s.userRepo.FindAdmins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
	user, err := s.userRepo.FindById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

// GetIdentities возвращает все платформы, привязанные к аккаунту (после
// ResolveAndLink их может быть больше одной — например telegram + vk).
func (s *UserService) GetIdentities(ctx context.Context, userID int) ([]Identity, error) {
	identities, err := s.identityRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list identities: %w", err)
	}
	return identities, nil
}

// UpdateProfile меняет имя/фамилию targetID. Админ может менять кому угодно
// и когда угодно; обычный юзер — только себя, и только пока не одобрен
// (после ApproveUser имя "замораживается" — дальше только через админа).
func (s *UserService) UpdateProfile(ctx context.Context, targetID int, firstName, lastName string, requester *http2.Claims) (*User, error) {
	user, err := s.userRepo.FindById(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if !requester.IsAdmin {
		if requester.UserID != targetID {
			return nil, &apperrors.ForbiddenError{Reason: "можно менять только свой профиль"}
		}
		if user.IsApproved {
			return nil, &apperrors.ForbiddenError{Reason: "профиль уже одобрен, изменить имя может только администратор"}
		}
	}

	user.FirstName = firstName
	user.LastName = lastName
	// Любое успешное сохранение имени — и от самого юзера с обязательного
	// экрана регистрации, и от админа — закрывает шаг регистрации.
	user.IsRegistered = true

	updated, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updated, nil
}

func (service *UserService) ApproveUser(ctx context.Context, id int) error {
	user, err := service.userRepo.FindById(ctx, id)

	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	user.IsApproved = true

	_, err = service.userRepo.Update(ctx, user)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (service *UserService) UnapproveUser(ctx context.Context, id int) error {
	user, err := service.userRepo.FindById(ctx, id)

	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	user.IsApproved = false

	_, err = service.userRepo.Update(ctx, user)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (service *UserService) DeleteUser(ctx context.Context, id int) error {
	user, err := service.userRepo.FindById(ctx, id)

	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	err = service.userRepo.Delete(ctx, user)

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *UserService) MakeAdmin(ctx context.Context, id int) error {
	user, err := s.userRepo.FindById(ctx, id)

	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user.IsAdmin {
		return &apperrors.ConflictError{Resource: fmt.Sprintf("user %d", id), Reason: "already an admin"}
	}

	user.IsAdmin = true

	_, err = s.userRepo.Update(ctx, user)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *UserService) RemoveAdmin(ctx context.Context, id int) error {
	user, err := s.userRepo.FindById(ctx, id)

	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if !user.IsAdmin {
		return &apperrors.ConflictError{Resource: fmt.Sprintf("user %d", id), Reason: "not an admin"}
	}

	user.IsAdmin = false

	_, err = s.userRepo.Update(ctx, user)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
