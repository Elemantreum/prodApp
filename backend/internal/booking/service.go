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
	FindAllByTrainingId(ctx context.Context, trainingId int) ([]Booking, error)
	GetAll(ctx context.Context) ([]Booking, error)
	GetAntiLeaderboard(ctx context.Context) ([]SinnerStats, error)
}

type BookingService struct {
	bookRepo     BookingRepository
	trainingRepo training.TrainingRepository
	userRepo     user.UserRepository
	identityRepo user.IdentityRepository
}

func NewBookingService(bookRepo BookingRepository, trainingRepo training.TrainingRepository, userRepo user.UserRepository, identityRepo user.IdentityRepository) *BookingService {
	return &BookingService{bookRepo: bookRepo, trainingRepo: trainingRepo, userRepo: userRepo, identityRepo: identityRepo}
}

func (service *BookingService) CreateBooking(ctx context.Context, userId int, trainingId int) (*Booking, error) {

	u, err := service.userRepo.FindById(ctx, userId)

	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if !u.IsApproved {
		return nil, &apperrors.ForbiddenError{Reason: "аккаунт ещё не одобрен администратором"}
	}

	tr, err := service.trainingRepo.FindById(ctx, trainingId)

	if err != nil {
		return nil, fmt.Errorf("training with id %v not found: %w", trainingId, err)
	}

	if !tr.IsActive {
		return nil, &apperrors.ConflictError{Resource: fmt.Sprintf("training %d", trainingId), Reason: "тренировка неактивна"}
	}

	if tr.Date.Before(time.Now()) {
		return nil, &apperrors.ConflictError{Resource: fmt.Sprintf("training %d", trainingId), Reason: "тренировка уже прошла"}
	}

	if tr.BookedCount >= tr.MaxQuantity {
		return nil, &apperrors.ConflictError{Resource: fmt.Sprintf("training %d", trainingId), Reason: "нет свободных мест"}
	}

	var booking Booking = Booking{UserId: userId, TrainingId: trainingId}

	saved, err := service.bookRepo.Create(ctx, &booking)

	if err != nil {
		return nil, fmt.Errorf("error creating booking: %w", err)
	}

	return saved, nil
}

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

		identities, err := service.identityRepo.FindByUserID(ctx, b.UserId)

		if err != nil {
			return nil, fmt.Errorf("error getting identities for user %d: %w", b.UserId, err)
		}

		participants = append(participants, Participant{
			UserId:      b.UserId,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			BookingDate: b.BookingDate,
			Identities:  identities,
		})
	}

	return participants, nil
}

func (service *BookingService) GetTrainingBookings(ctx context.Context, trainingId int) ([]BookingDetail, error) {
	_, err := service.trainingRepo.FindById(ctx, trainingId)

	if err != nil {
		return nil, fmt.Errorf("training with id %v not found: %w", trainingId, err)
	}

	bookings, err := service.bookRepo.FindAllByTrainingId(ctx, trainingId)

	if err != nil {
		return nil, fmt.Errorf("error getting bookings: %w", err)
	}

	details := make([]BookingDetail, 0, len(bookings))

	for _, b := range bookings {
		u, err := service.userRepo.FindById(ctx, b.UserId)

		if err != nil {
			return nil, fmt.Errorf("error getting user %d: %w", b.UserId, err)
		}

		identities, err := service.identityRepo.FindByUserID(ctx, b.UserId)

		if err != nil {
			return nil, fmt.Errorf("error getting identities for user %d: %w", b.UserId, err)
		}

		details = append(details, BookingDetail{
			BookingId:   b.Id,
			UserId:      b.UserId,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			Status:      b.Status,
			BookingDate: b.BookingDate,
			Identities:  identities,
		})
	}

	return details, nil
}

func (service *BookingService) SetAttendance(ctx context.Context, bookingId int, attended bool) (*Booking, error) {
	b, err := service.bookRepo.FindById(ctx, bookingId)

	if err != nil {
		return nil, fmt.Errorf("error getting booking: %w", err)
	}

	if b.Status == Cancelled {
		return nil, &apperrors.ConflictError{Resource: fmt.Sprintf("booking with id %v", bookingId), Reason: "нельзя отметить посещаемость отменённой записи"}
	}

	if attended {
		b.Status = Completed
	} else {
		b.Status = NotCompleted
	}

	err = service.bookRepo.Update(ctx, b)

	if err != nil {
		return nil, fmt.Errorf("error updating booking: %w", err)
	}

	return b, nil
}

func (service *BookingService) HasBookings(ctx context.Context, userId int) (bool, error) {
	bookings, err := service.bookRepo.FindByUserId(ctx, userId)

	if err != nil {
		return false, fmt.Errorf("error checking bookings: %w", err)
	}

	return len(bookings) > 0, nil
}

func (service *BookingService) GetAntiLeaderboard(ctx context.Context) ([]SinnerStats, error) {
	sinners, err := service.bookRepo.GetAntiLeaderboard(ctx)

	if err != nil {
		return nil, fmt.Errorf("error getting anti leaderboard: %w", err)
	}

	return sinners, nil
}

func (service *BookingService) GetAllBookingsByUserId(ctx context.Context, userId int, requester *http2.Claims) ([]Booking, error) {
	if !requester.IsAdmin && userId != requester.UserID {
		return nil, &apperrors.ForbiddenError{}
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
