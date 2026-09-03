package booking

import (
	"context"
	"fmt"
	"prodApp/internal/platform/apperrors"
	http2 "prodApp/internal/platform/http"
	"prodApp/internal/training"
	"prodApp/internal/user"
	"time"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *Booking) (*Booking, error)
	Update(ctx context.Context, booking *Booking) error
	Delete(ctx context.Context, id int) error
	FindById(ctx context.Context, id int) (*Booking, error)
	FindByUserId(ctx context.Context, userId int) ([]Booking, error)
	FindByTrainingId(ctx context.Context, trainingId int) ([]Booking, error)
	GetAll(ctx context.Context) ([]Booking, error)
}

// Participant — то, что видит обычный пользователь про других записавшихся:
// имя, ник и когда записался. Никакого approve/admin статуса — это не его
// дело, в отличие от админской ручки /user/get.
type Participant struct {
	UserId      int       `json:"userId"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Username    string    `json:"username"`
	BookingDate time.Time `json:"bookingDate"`
}

type BookingService struct {
	bookRepo     BookingRepository
	trainingRepo training.TrainingRepository
	userRepo     user.UserRepository
}

func NewBookingService(bookRepo BookingRepository, trainingRepo training.TrainingRepository, userRepo user.UserRepository) *BookingService {
	return &BookingService{bookRepo: bookRepo, trainingRepo: trainingRepo, userRepo: userRepo}
}

func (service *BookingService) CreateBooking(ctx context.Context, userId int, trainingId int) (*Booking, error) {

	u, err := service.userRepo.FindById(ctx, userId)

	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Одобрение — не декоративное поле: пока админ не одобрил юзера, тот не
	// может записываться на тренировки. Проверяем по свежим данным из БД, а
	// не по claims из JWT — иначе одобрение админом не сработает мгновенно,
	// а только после того как у юзера обновится токен (до 15 минут).
	if !u.IsApproved {
		return nil, &apperrors.ForbiddenError{Reason: "аккаунт ещё не одобрен администратором"}
	}

	_, err = service.trainingRepo.FindById(ctx, trainingId)

	if err != nil {
		return nil, fmt.Errorf("training with id %v not found: %w", trainingId, err)
	}

	var booking Booking = Booking{UserId: userId, TrainingId: trainingId}

	saved, err := service.bookRepo.Create(ctx, &booking)

	if err != nil {
		return nil, fmt.Errorf("error creating booking: %w", err)
	}

	return saved, nil
}

// GetParticipants — кто сейчас записан на тренировку. Доступно любому
// авторизованному пользователю (не только админам) — это публичный список
// внутри приложения, а не приватные данные конкретного юзера.
func (service *BookingService) GetParticipants(ctx context.Context, trainingId int) ([]Participant, error) {
	_, err := service.trainingRepo.FindById(ctx, trainingId)

	if err != nil {
		return nil, fmt.Errorf("training with id %v not found: %w", trainingId, err)
	}

	bookings, err := service.bookRepo.FindByTrainingId(ctx, trainingId)

	if err != nil {
		return nil, fmt.Errorf("error getting bookings: %w", err)
	}

	participants := make([]Participant, 0, len(bookings))

	for _, b := range bookings {
		u, err := service.userRepo.FindById(ctx, b.UserId)

		if err != nil {
			return nil, fmt.Errorf("error getting user %d: %w", b.UserId, err)
		}

		participants = append(participants, Participant{
			UserId:      b.UserId,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			Username:    u.Username,
			BookingDate: b.BookingDate,
		})
	}

	return participants, nil
}

func (service *BookingService) GetAllBookingsByUserId(ctx context.Context, userId int, requester *http2.Claims) ([]Booking, error) {
	if !requester.IsAdmin && userId != requester.UserID {
		return nil, apperrors.ForbiddenError{}
	}

	bookings, err := service.bookRepo.FindByUserId(ctx, userId)

	if err != nil {
		return nil, fmt.Errorf("error getting bookings: %w", err)
	}

	return bookings, nil
}

func (service BookingService) CancelBooking(ctx context.Context, bookingId int, requester *http2.Claims) error {
	booking, err := service.bookRepo.FindById(ctx, bookingId)

	if err != nil {
		return fmt.Errorf("error getting booking: %w", err)
	}

	if !requester.IsAdmin && booking.UserId != requester.UserID {
		return &apperrors.ForbiddenError{}
	}

	if booking.UnbookingDate != nil {
		return &apperrors.ConflictError{Resource: fmt.Sprintf("booking with id %v", bookingId), Reason: "already unbooked"}
	}

	now := time.Now()
	booking.Status = Cancelled
	booking.UnbookingDate = &now

	err = service.bookRepo.Update(ctx, booking)

	if err != nil {
		return fmt.Errorf("error updating booking: %w", err)
	}

	return nil
}
