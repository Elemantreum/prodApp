package booking

import (
	"context"
	"fmt"
	"prodApp/internal/platform/apperrors"
	"prodApp/internal/section"
	"prodApp/internal/training"
	"prodApp/internal/user"
	"time"
)

type BookingRepository interface {
	History(ctx context.Context, sectionID, userID int) ([]HistoryItem, error)
	WaitingMembers(ctx context.Context, trainingID int) ([]WaitingMember, error)
	WithLockedTrainings(ctx context.Context, source, target int, apply func(RosterRepository) error) error
	Waitlist(ctx context.Context, userID int) ([]int, error)
	LeaveWaitlist(ctx context.Context, userID, trainingID int) error
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

type RosterRepository interface {
	Availability(ctx context.Context, trainingID int) (free int, open, past bool, err error)
	IsRegisteredMember(ctx context.Context, userID int) (bool, error)
	HasRosterEntry(ctx context.Context, userID, trainingID, exceptBookingID int) (bool, error)
	WaitlistEligibility(ctx context.Context, userID, trainingID int) (eligible, booked bool, err error)
	JoinWaitlist(ctx context.Context, userID, trainingID int) error
	AddBooking(ctx context.Context, userID, trainingID int) error
	MoveBooking(ctx context.Context, bookingID, sourceID, targetID, userID int) error
	DeleteBooking(ctx context.Context, bookingID, trainingID int) error
	ResetAttendance(ctx context.Context, bookingID, trainingID int) error
	AddWaitingMember(ctx context.Context, userID, trainingID int) error
	RemoveWaitingMember(ctx context.Context, userID, trainingID int) error
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

	if !u.IsRegistered {
		return nil, &apperrors.ForbiddenError{Reason: "сначала заполните профиль"}
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

	booking := Booking{UserId: userId, TrainingId: trainingId}

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

func (service *BookingService) GetAllBookingsByUserId(ctx context.Context, userId int) ([]Booking, error) {
	bookings, err := service.bookRepo.FindByUserId(ctx, userId)

	if err != nil {
		return nil, fmt.Errorf("error getting bookings: %w", err)
	}

	return bookings, nil
}

func (service *BookingService) GetBooking(ctx context.Context, bookingId int) (*Booking, error) {
	booking, err := service.bookRepo.FindById(ctx, bookingId)
	if err != nil {
		return nil, fmt.Errorf("error getting booking: %w", err)
	}
	return booking, nil
}

func (service *BookingService) CancelBooking(ctx context.Context, booking Booking) error {
	if booking.UnbookingDate != nil {
		return &apperrors.ConflictError{Resource: fmt.Sprintf("booking with id %v", booking.Id), Reason: "already unbooked"}
	}

	now := time.Now()
	booking.Status = Cancelled
	booking.UnbookingDate = &now

	err := service.bookRepo.Update(ctx, &booking)

	if err != nil {
		return fmt.Errorf("error updating booking: %w", err)
	}

	return nil
}

func (service *BookingService) OwnWaitlist(ctx context.Context, userID int) ([]int, error) {
	return service.bookRepo.Waitlist(ctx, userID)
}

func (service *BookingService) SetWaiting(ctx context.Context, userID, trainingID int, join bool) error {
	if !join {
		return service.bookRepo.LeaveWaitlist(ctx, userID, trainingID)
	}
	return service.bookRepo.WithLockedTrainings(ctx, trainingID, trainingID, func(roster RosterRepository) error {
		free, open, _, err := roster.Availability(ctx, trainingID)
		if err != nil {
			return err
		}
		if !open {
			return &apperrors.ConflictError{Resource: "тренировка", Reason: "запись закрыта"}
		}
		if free > 0 {
			return &apperrors.ConflictError{Resource: "тренировка", Reason: "есть свободные места"}
		}
		eligible, booked, err := roster.WaitlistEligibility(ctx, userID, trainingID)
		if err != nil {
			return err
		}
		if !eligible {
			return &apperrors.ForbiddenError{Reason: "нужно заполнить профиль и иметь доступ к секции"}
		}
		if booked {
			return &apperrors.ConflictError{Resource: "тренировка", Reason: "уже есть запись"}
		}
		return roster.JoinWaitlist(ctx, userID, trainingID)
	})
}

func (service *BookingService) Manage(ctx context.Context, change BookingChange) error {
	if change.TargetID == 0 {
		change.TargetID = change.TrainingID
	}
	return service.bookRepo.WithLockedTrainings(ctx, change.TrainingID, change.TargetID, func(roster RosterRepository) error {
		switch change.Action {
		case "create", "edit", "wait":
			return service.placeMember(ctx, roster, change)
		case "reset":
			free, _, past, err := roster.Availability(ctx, change.TrainingID)
			if err != nil {
				return err
			}
			if !past && free <= 0 {
				return &apperrors.ConflictError{Resource: "тренировка", Reason: "нет свободных мест"}
			}
			return roster.ResetAttendance(ctx, change.BookingID, change.TrainingID)
		case "delete":
			return roster.DeleteBooking(ctx, change.BookingID, change.TrainingID)
		case "unwait":
			return roster.RemoveWaitingMember(ctx, change.UserID, change.TargetID)
		default:
			return &apperrors.ValidationError{Field: "action"}
		}
	})
}

func (service *BookingService) placeMember(ctx context.Context, roster RosterRepository, change BookingChange) error {
	member, err := roster.IsRegisteredMember(ctx, change.UserID)
	if err != nil {
		return err
	}
	if !member {
		return &apperrors.ConflictError{Resource: "участник", Reason: "нужен зарегистрированный участник секции"}
	}
	free, open, past, err := roster.Availability(ctx, change.TargetID)
	if err != nil {
		return err
	}
	if !open && (!past || change.Action == "wait") {
		return &apperrors.ConflictError{Resource: "тренировка", Reason: "запись закрыта"}
	}
	if change.Action == "wait" && free > 0 {
		return &apperrors.ConflictError{Resource: "тренировка", Reason: "есть свободные места"}
	}
	if free <= 0 && (change.Action == "create" || change.TargetID != change.TrainingID) {
		return &apperrors.ConflictError{Resource: "тренировка", Reason: "нет свободных мест"}
	}
	if past {
		var exceptBookingID int
		if change.Action == "edit" {
			exceptBookingID = change.BookingID
		}
		existing, err := roster.HasRosterEntry(ctx, change.UserID, change.TargetID, exceptBookingID)
		if err != nil {
			return err
		}
		if existing {
			return &apperrors.ConflictError{Resource: "участник", Reason: "уже записан"}
		}
	}
	switch change.Action {
	case "create":
		return roster.AddBooking(ctx, change.UserID, change.TargetID)
	case "edit":
		return roster.MoveBooking(ctx, change.BookingID, change.TrainingID, change.TargetID, change.UserID)
	default:
		if change.PreviousUserID != 0 {
			if err := roster.RemoveWaitingMember(ctx, change.PreviousUserID, change.TargetID); err != nil {
				return err
			}
		}
		return roster.AddWaitingMember(ctx, change.UserID, change.TargetID)
	}
}

func (service *BookingService) WaitingMembers(ctx context.Context, trainingID int) ([]WaitingMember, error) {
	return service.bookRepo.WaitingMembers(ctx, trainingID)
}

func (service *BookingService) SectionBookings(ctx context.Context) ([]HistoryItem, error) {
	return service.bookRepo.History(ctx, section.ID(ctx), 0)
}

func (service *BookingService) UserHistory(ctx context.Context, userID int) ([]HistoryItem, error) {
	return service.bookRepo.History(ctx, 0, userID)
}
