package user

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"prodApp/internal/platform/apperrors"
	"prodApp/internal/platform/outbox"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	UpdateProfile(ctx context.Context, user *User) (*User, error)
	DeleteOrphan(ctx context.Context, userID int) error
	Delete(ctx context.Context, user *User) error
	FindById(ctx context.Context, id int) (*User, error)
	GetAll(ctx context.Context) ([]User, error)
	FindAdmins(ctx context.Context) ([]User, error)
}
type IdentityRepository interface {
	Create(ctx context.Context, identity *Identity) (*Identity, error)

	FindByProviderAndExternalID(ctx context.Context, provider, externalID string) (*Identity, error)
	FindByUserID(ctx context.Context, userID int) ([]Identity, error)
	Reassign(ctx context.Context, provider, externalID, username string, newUserID int) error
	GetBatchById(ctx context.Context, ids []int, provider string) ([]Identity, error)
	UpdateUsername(ctx context.Context, provider, externalID, username string) error
	AdminTelegramChatIds(ctx context.Context) ([]int64, error)
}

type LinkCodeRepository interface {
	Create(ctx context.Context, code *LinkCode) error
	FindAndDelete(ctx context.Context, code string) (*LinkCode, error)
}

type OnNewUser struct {
	UserId     int
	FirstName  string
	LastName   string
	Provider   string
	ExternalId string

	ChatIds []int64
}

type UserService struct {
	userRepo     UserRepository
	identityRepo IdentityRepository
	linkCodeRepo LinkCodeRepository
	outbox       *outbox.Repository
}

func NewUserService(userRepo UserRepository, identityRepo IdentityRepository, linkCodeRepo LinkCodeRepository, outboxRepo *outbox.Repository) *UserService {
	return &UserService{userRepo: userRepo, identityRepo: identityRepo, linkCodeRepo: linkCodeRepo, outbox: outboxRepo}
}

const (
	linkCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	linkCodeLength   = 8
	linkCodeTTL      = 5 * time.Minute
)

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

func (s *UserService) ResolveAndLink(ctx context.Context, code, provider, externalID, username string) (*User, error) {
	lc, err := s.linkCodeRepo.FindAndDelete(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("invalid link code: %w", err)
	}

	if time.Now().After(lc.ExpiresAt) {
		return nil, &apperrors.ValidationError{Field: "code: истёк, запросите новый"}
	}

	var previousOwnerID int
	if existing, err := s.identityRepo.FindByProviderAndExternalID(ctx, provider, externalID); err == nil {
		previousOwnerID = existing.UserId
	}

	if err := s.identityRepo.Reassign(ctx, provider, externalID, username, lc.UserId); err != nil {
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
}

func (s *UserService) cleanupOrphan(ctx context.Context, userID int) {
	if err := s.userRepo.DeleteOrphan(ctx, userID); err != nil {
		log.Printf("delete orphan user %d: %v", userID, err)
	}
}

func (s *UserService) EnsureProfile(ctx context.Context, provider, externalID, firstName, lastName, username string) (*User, error) {
	identity, err := s.identityRepo.FindByProviderAndExternalID(ctx, provider, externalID)
	if err == nil {
		if updateErr := s.identityRepo.UpdateUsername(ctx, provider, externalID, username); updateErr != nil {
			return nil, fmt.Errorf("update identity username: %w", updateErr)
		}
		return s.userRepo.FindById(ctx, identity.UserId)
	}
	if !errors.Is(err, apperrors.NotFoundError{}) {
		return nil, fmt.Errorf("find identity: %w", err)
	}

	newUser := &User{
		FirstName: firstName,
		LastName:  lastName,
		IsAdmin:   false,

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
		Username:   username,
	})
	if err != nil {
		return nil, fmt.Errorf("create identity: %w", err)
	}

	adminChatIds, err := s.identityRepo.AdminTelegramChatIds(ctx)
	if err != nil {

		log.Printf("resolve admin chat ids: %v", err)
	}

	event := OnNewUser{
		UserId:     createdUser.Id,
		FirstName:  createdUser.FirstName,
		LastName:   createdUser.LastName,
		Provider:   provider,
		ExternalId: externalID,
		ChatIds:    adminChatIds,
	}

	if err := outbox.Create(ctx, s.outbox, event); err != nil {
		log.Printf("outbox: %v", err)
	}

	return createdUser, nil
}

func (s *UserService) LinkIdentity(ctx context.Context, userID int, provider, externalID string) error {
	existing, err := s.identityRepo.FindByProviderAndExternalID(ctx, provider, externalID)
	if err == nil {
		if existing.UserId != userID {
			return &apperrors.ConflictError{
				Resource: "identity",
				Reason:   "already linked to a different account",
			}
		}
		return nil
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

func (s *UserService) GetIdentities(ctx context.Context, userID int) ([]Identity, error) {
	identities, err := s.identityRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list identities: %w", err)
	}
	return identities, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, user User, firstName, lastName string) (*User, error) {
	user.FirstName = firstName
	user.LastName = lastName

	user.IsRegistered = true

	updated, err := s.userRepo.UpdateProfile(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updated, nil
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
