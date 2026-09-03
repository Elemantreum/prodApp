package training

import (
	"context"
	"fmt"
	"time"
)

type TrainingRepository interface {
	Create(ctx context.Context, description string, date time.Time, maxQuantity int, duration time.Duration) (*Training, error)
	Update(ctx context.Context, id int, description string, date time.Time, maxQuantity int, duration time.Duration, isActive bool) error
	Delete(ctx context.Context, id int) error
	FindById(ctx context.Context, id int) (*Training, error)
	GetAll(ctx context.Context) ([]Training, error)
}

type TrainingService struct {
	repo TrainingRepository
}

func NewTrainingService(repo TrainingRepository) *TrainingService {
	return &TrainingService{repo: repo}
}

func (service *TrainingService) AddTraining(ctx context.Context, description string, date time.Time, maxQuantity int, duration time.Duration) (*Training, error) {
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

	saved, err := service.repo.Create(ctx, description, date, maxQuantity, duration)

	if err != nil {
		return nil, fmt.Errorf("saving user: %w", err)
	}

	return saved, nil
}

func (service *TrainingService) UpdateTraining(ctx context.Context, id int, description string, date time.Time, maxQuantity int, duration time.Duration, isActive bool) (*Training, error) {
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

	err := service.repo.Update(ctx, id, description, date, maxQuantity, duration, isActive)

	if err != nil {
		return nil, fmt.Errorf("updating training: %w", err)
	}

	updated, err := service.repo.FindById(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("finding updated training: %w", err)
	}

	return updated, nil
}

func (service *TrainingService) FindById(ctx context.Context, id int) (*Training, error) {
	training, err := service.repo.FindById(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}

	return training, nil
}

func (service *TrainingService) GetAll(ctx context.Context) ([]Training, error) {
	trainings, err := service.repo.GetAll(ctx)

	if err != nil {
		return nil, fmt.Errorf("getting trainings: %w", err)
	}

	return trainings, nil
}
