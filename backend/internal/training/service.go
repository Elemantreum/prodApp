package training

import (
	"context"
	"fmt"
	"log"
	"prodApp/internal/platform/apperrors"
	http2 "prodApp/internal/platform/http"
	"prodApp/internal/platform/outbox"
	"time"
)

type TrainingRepository interface {
	Create(ctx context.Context, description string, date time.Time, maxQuantity int, duration time.Duration) (*Training, error)
	Update(ctx context.Context, id int, description string, date time.Time, maxQuantity int, duration time.Duration, isActive bool) error
	Delete(ctx context.Context, id int) error
	FindById(ctx context.Context, id int) (*Training, error)
	GetAll(ctx context.Context) ([]Training, error)

	CreateFromOccurrence(ctx context.Context, recurringId int, description string, date time.Time, maxQuantity int, duration time.Duration) (*Training, error)

	HasBookings(ctx context.Context, trainingId int) (bool, error)
	DeactivatePast(ctx context.Context) error
}

type RecurringTrainingRepository interface {
	Create(ctx context.Context, description string, anchorDate time.Time, maxQuantity int, duration time.Duration, leadTime time.Duration) (*RecurringTraining, error)
	Update(ctx context.Context, id int, description string, anchorDate time.Time, maxQuantity int, duration time.Duration, leadTime time.Duration, isActive bool) (*RecurringTraining, error)
	GetAll(ctx context.Context) ([]RecurringTraining, error)
	FindActive(ctx context.Context) ([]RecurringTraining, error)
}

type ChatIdResolver interface {
	AllTelegramChatIds(ctx context.Context) ([]int64, error)
}

type TrainingService struct {
	repo          TrainingRepository
	recurringRepo RecurringTrainingRepository
	outbox        *outbox.Repository
	chatIds       ChatIdResolver
}

func NewTrainingService(repo TrainingRepository, recurringRepo RecurringTrainingRepository, outboxRepo *outbox.Repository, chatIds ChatIdResolver) *TrainingService {
	return &TrainingService{repo: repo, recurringRepo: recurringRepo, outbox: outboxRepo, chatIds: chatIds}
}

func (s *TrainingService) AddTraining(ctx context.Context, description string, date time.Time, maxQuantity int, duration time.Duration) (*Training, error) {
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}

	if date.IsZero() {
		return nil, fmt.Errorf("date is required")
	}

	if maxQuantity <= 0 {
		return nil, fmt.Errorf("maxQuantity can't be less or equal than 0")
	}

	if duration <= 0 {
		return nil, fmt.Errorf("duration can't be less or equal than 0")
	}

	saved, err := s.repo.Create(ctx, description, date, maxQuantity, duration)

	if err != nil {
		return nil, fmt.Errorf("saving training: %w", err)
	}

	s.publishCreateEvent(ctx, saved)

	return saved, nil
}

func (s *TrainingService) publishCreateEvent(ctx context.Context, t *Training) {
	chatIds, err := s.chatIds.AllTelegramChatIds(ctx)
	if err != nil {

		log.Printf("resolve chat ids: %v", err)
	}

	event := OnCreateTraining{
		TrainingId:  t.Id,
		Description: t.Description,
		Date:        t.Date,
		MaxQuantity: t.MaxQuantity,
		Duration:    t.Duration,
		ChatIds:     chatIds,
	}

	if err := outbox.Create(ctx, s.outbox, event); err != nil {
		log.Printf("outbox: %v", err)
	}
}

func (s *TrainingService) UpdateTraining(ctx context.Context, id int, description string, date time.Time, maxQuantity int, duration time.Duration, isActive bool) (*Training, error) {
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}

	if date.IsZero() {
		return nil, fmt.Errorf("date is required")
	}

	if maxQuantity <= 0 {
		return nil, fmt.Errorf("maxQuantity can't be less or equal than 0")
	}

	if duration <= 0 {
		return nil, fmt.Errorf("duration can't be less or equal than 0")
	}

	before, err := s.repo.FindById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding training before update: %w", err)
	}

	err = s.repo.Update(ctx, id, description, date, maxQuantity, duration, isActive)

	if err != nil {
		return nil, fmt.Errorf("updating training: %w", err)
	}

	updated, err := s.repo.FindById(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("finding updated training: %w", err)
	}

	chatIds, err := s.chatIds.AllTelegramChatIds(ctx)
	if err != nil {
		log.Printf("resolve chat ids: %v", err)
	}

	event := OnUpdateTraining{
		TrainingId:     updated.Id,
		ChatIds:        chatIds,
		Description:    updated.Description,
		OldDescription: before.Description,
		Date:           updated.Date,
		OldDate:        before.Date,
		MaxQuantity:    updated.MaxQuantity,
		OldMaxQuantity: before.MaxQuantity,
		Duration:       updated.Duration,
		OldDuration:    before.Duration,
		IsActive:       updated.IsActive,
		OldIsActive:    before.IsActive,
	}

	if err := outbox.Create(ctx, s.outbox, event); err != nil {
		log.Printf("outbox: %v", err)
	}

	return updated, nil
}

func (s *TrainingService) FindById(ctx context.Context, id int) (*Training, error) {
	training, err := s.repo.FindById(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}

	return training, nil
}

func (s *TrainingService) GetAll(ctx context.Context, requester *http2.Claims) ([]Training, error) {
	trainings, err := s.repo.GetAll(ctx)

	if err != nil {
		return nil, fmt.Errorf("getting trainings: %w", err)
	}

	if requester != nil && requester.IsAdmin {
		return trainings, nil
	}

	active := make([]Training, 0, len(trainings))
	for _, t := range trainings {
		if t.IsActive {
			active = append(active, t)
		}
	}

	return active, nil
}

func (s *TrainingService) DeleteTraining(ctx context.Context, id int) error {
	hasBookings, err := s.repo.HasBookings(ctx, id)

	if err != nil {
		return fmt.Errorf("checking bookings: %w", err)
	}

	if hasBookings {
		return &apperrors.ConflictError{
			Resource: fmt.Sprintf("training %d", id),
			Reason:   "нельзя удалить тренировку, на которую есть записи — сначала деактивируйте её",
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting training: %w", err)
	}

	return nil
}

func (s *TrainingService) DeactivatePastTrainings(ctx context.Context) error {
	if err := s.repo.DeactivatePast(ctx); err != nil {
		return fmt.Errorf("deactivating past trainings: %w", err)
	}

	return nil
}

const week = 7 * 24 * time.Hour

func (service *TrainingService) CreateRule(ctx context.Context, description string, anchorDate time.Time, maxQuantity int, duration time.Duration, leadTime time.Duration) (*RecurringTraining, error) {
	if err := validateRule(description, anchorDate, maxQuantity, duration, leadTime); err != nil {
		return nil, err
	}

	saved, err := service.recurringRepo.Create(ctx, description, anchorDate, maxQuantity, duration, leadTime)

	if err != nil {
		return nil, fmt.Errorf("saving recurring training: %w", err)
	}

	if err := service.generateForRule(ctx, saved); err != nil {
		return nil, fmt.Errorf("generating initial occurrences: %w", err)
	}

	return saved, nil
}

func (service *TrainingService) ListRules(ctx context.Context) ([]RecurringTraining, error) {
	rules, err := service.recurringRepo.GetAll(ctx)

	if err != nil {
		return nil, fmt.Errorf("listing recurring trainings: %w", err)
	}

	return rules, nil
}

func (service *TrainingService) UpdateRule(ctx context.Context, id int, description string, anchorDate time.Time, maxQuantity int, duration time.Duration, leadTime time.Duration, isActive bool) (*RecurringTraining, error) {
	if err := validateRule(description, anchorDate, maxQuantity, duration, leadTime); err != nil {
		return nil, err
	}

	updated, err := service.recurringRepo.Update(ctx, id, description, anchorDate, maxQuantity, duration, leadTime, isActive)

	if err != nil {
		return nil, fmt.Errorf("updating recurring training: %w", err)
	}

	if updated.IsActive {
		if err := service.generateForRule(ctx, updated); err != nil {
			return nil, fmt.Errorf("generating occurrences after update: %w", err)
		}
	}

	return updated, nil
}

func validateRule(description string, anchorDate time.Time, maxQuantity int, duration time.Duration, leadTime time.Duration) error {
	if description == "" {
		return fmt.Errorf("description is required")
	}

	if anchorDate.IsZero() {
		return fmt.Errorf("anchorDate is required")
	}

	if maxQuantity <= 0 {
		return fmt.Errorf("maxQuantity can't be less or equal than 0")
	}

	if duration <= 0 {
		return fmt.Errorf("duration can't be less or equal than 0")
	}

	if leadTime < 0 {
		return fmt.Errorf("leadTime can't be negative")
	}

	return nil
}

func (service *TrainingService) GenerateOccurrences(ctx context.Context) error {
	rules, err := service.recurringRepo.FindActive(ctx)

	if err != nil {
		return fmt.Errorf("listing active recurring trainings: %w", err)
	}

	for _, rule := range rules {
		if err := service.generateForRule(ctx, &rule); err != nil {
			return fmt.Errorf("generating occurrences for rule %d: %w", rule.Id, err)
		}
	}

	return nil
}

func (service *TrainingService) generateForRule(ctx context.Context, rule *RecurringTraining) error {
	now := time.Now()

	occurrence := rule.AnchorDate
	for occurrence.Sub(now) <= rule.LeadTime {
		created, err := service.repo.CreateFromOccurrence(ctx, rule.Id, rule.Description, occurrence, rule.MaxQuantity, rule.Duration)
		if err != nil {
			return err
		}
		if created != nil {
			service.publishCreateEvent(ctx, created)
		}
		occurrence = occurrence.Add(week)
	}

	return nil
}
