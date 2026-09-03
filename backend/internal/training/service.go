package training

import (
	"context"
	"fmt"
	"prodApp/internal/platform/apperrors"
	http2 "prodApp/internal/platform/http"
	"time"
)

type TrainingRepository interface {
	Create(ctx context.Context, description string, date time.Time, maxQuantity int, duration time.Duration) (*Training, error)
	Update(ctx context.Context, id int, description string, date time.Time, maxQuantity int, duration time.Duration, isActive bool) error
	Delete(ctx context.Context, id int) error
	FindById(ctx context.Context, id int) (*Training, error)
	GetAll(ctx context.Context) ([]Training, error)
	CreateFromOccurrence(ctx context.Context, recurringId int, description string, date time.Time, maxQuantity int, duration time.Duration) error
	// HasBookings — прямой SQL-запрос к таблице bookings (без импорта пакета
	// booking — он уже импортирует training, обратная зависимость была бы
	// циклом). Используется, чтобы не дать снести тренировку, на которую уже
	// есть история записей/посещаемости — см. DeleteTraining.
	HasBookings(ctx context.Context, trainingId int) (bool, error)
	DeactivatePast(ctx context.Context) error
}

type RecurringTrainingRepository interface {
	Create(ctx context.Context, description string, anchorDate time.Time, maxQuantity int, duration time.Duration, leadTime time.Duration) (*RecurringTraining, error)
	Update(ctx context.Context, id int, description string, anchorDate time.Time, maxQuantity int, duration time.Duration, leadTime time.Duration, isActive bool) (*RecurringTraining, error)
	GetAll(ctx context.Context) ([]RecurringTraining, error)
	FindActive(ctx context.Context) ([]RecurringTraining, error)
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

// GetAll возвращает все тренировки админу и только активные — всем
// остальным. requester может быть nil (на всякий случай, если middleware
// почему-то не положил claims в контекст) — тогда, как и для любого
// обычного пользователя, неактивные тренировки не отдаются.
func (service *TrainingService) GetAll(ctx context.Context, requester *http2.Claims) ([]Training, error) {
	trainings, err := service.repo.GetAll(ctx)

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

// DeleteTraining удаляет тренировку насовсем. Отказывает, если на неё есть
// хоть одна бронь (активная, отменённая, отмеченная посещаемостью — любая):
// это уже история, каскадное удаление через FK молча снесло бы её вместе с
// тренировкой. Деактивация (UpdateTraining с isActive=false) — правильный
// способ "убрать" тренировку с сохранением истории; Delete — только для
// пустых, ничего не значащих записей (опечатка, лишнее регулярное занятие
// и т.п.).
func (service *TrainingService) DeleteTraining(ctx context.Context, id int) error {
	hasBookings, err := service.repo.HasBookings(ctx, id)

	if err != nil {
		return fmt.Errorf("checking bookings: %w", err)
	}

	if hasBookings {
		return &apperrors.ConflictError{
			Resource: fmt.Sprintf("training %d", id),
			Reason:   "нельзя удалить тренировку, на которую есть записи — сначала деактивируйте её",
		}
	}

	if err := service.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting training: %w", err)
	}

	return nil
}

// DeactivatePastTrainings — вызывается периодически фоновым тикером (см.
// main.go): переводит прошедшие тренировки в isActive=false, чтобы они сами
// переставали быть видны обычным пользователям (см. GetAll), без ручного
// вмешательства админа. Booking.CreateBooking всё равно дополнительно
// проверяет дату напрямую — на случай гонки в окне между тем, как
// тренировка прошла, и следующим тиком этого метода.
func (service *TrainingService) DeactivatePastTrainings(ctx context.Context) error {
	if err := service.repo.DeactivatePast(ctx); err != nil {
		return fmt.Errorf("deactivating past trainings: %w", err)
	}

	return nil
}

const week = 7 * 24 * time.Hour

type RecurringService struct {
	repo         RecurringTrainingRepository
	trainingRepo TrainingRepository
}

func NewRecurringService(repo RecurringTrainingRepository, trainingRepo TrainingRepository) *RecurringService {
	return &RecurringService{repo: repo, trainingRepo: trainingRepo}
}

func (service *RecurringService) CreateRule(ctx context.Context, description string, anchorDate time.Time, maxQuantity int, duration time.Duration, leadTime time.Duration) (*RecurringTraining, error) {
	if err := validateRule(description, anchorDate, maxQuantity, duration, leadTime); err != nil {
		return nil, err
	}

	saved, err := service.repo.Create(ctx, description, anchorDate, maxQuantity, duration, leadTime)

	if err != nil {
		return nil, fmt.Errorf("saving recurring training: %w", err)
	}

	// Сразу материализуем первые занятия по новому правилу, не дожидаясь
	// фонового тика — админ должен увидеть результат немедленно.
	if err := service.generateForRule(ctx, saved); err != nil {
		return nil, fmt.Errorf("generating initial occurrences: %w", err)
	}

	return saved, nil
}

func (service *RecurringService) ListRules(ctx context.Context) ([]RecurringTraining, error) {
	rules, err := service.repo.GetAll(ctx)

	if err != nil {
		return nil, fmt.Errorf("listing recurring trainings: %w", err)
	}

	return rules, nil
}

// UpdateRule меняет все поля правила разом, включая isActive — так
// включение/выключение регулярности и правка остальных полей (даты первого
// занятия, мест, длительности, времени создания) идут через одну и ту же
// ручку, а не отдельный toggle. Уже созданные тренировки не трогает — они
// остаются обычными строками в trainings, которые админ может отдельно
// отредактировать/деактивировать через обычный UpdateTraining. Если правило
// снова активно и изменения открыли новые "созревшие" occurrences (например
// сдвинули anchorDate назад или увеличили leadTime), они материализуются
// сразу же, не дожидаясь фонового тика.
func (service *RecurringService) UpdateRule(ctx context.Context, id int, description string, anchorDate time.Time, maxQuantity int, duration time.Duration, leadTime time.Duration, isActive bool) (*RecurringTraining, error) {
	if err := validateRule(description, anchorDate, maxQuantity, duration, leadTime); err != nil {
		return nil, err
	}

	updated, err := service.repo.Update(ctx, id, description, anchorDate, maxQuantity, duration, leadTime, isActive)

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

// GenerateOccurrences — идемпотентный проход по всем активным правилам:
// досоздаёт недостающие занятия, чьё окно создания уже наступило (см.
// generateForRule и поле LeadTime). Вызывается и по расписанию (см.
// main.go), и сразу при создании/изменении правила. Безопасно звать сколько
// угодно раз — INSERT ... ON CONFLICT DO NOTHING на стороне репозитория
// гарантирует отсутствие дублей.
func (service *RecurringService) GenerateOccurrences(ctx context.Context) error {
	rules, err := service.repo.FindActive(ctx)

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

// generateForRule материализует все occurrences правила, чьё "окно
// создания" уже открылось — то есть occurrence.Sub(now) <= rule.LeadTime.
// Occurrences идут по возрастанию с фиксированным недельным шагом, так что
// как только встретили ещё не созревший — все следующие тем более не
// созрели, можно останавливаться. Отстающих (например после простоя
// сервера) это тоже само подхватывает: цикл просто стартует от anchor и
// создаёт всё, что уже пора, сколько бы occurrences ни накопилось.
func (service *RecurringService) generateForRule(ctx context.Context, rule *RecurringTraining) error {
	now := time.Now()

	occurrence := rule.AnchorDate
	for occurrence.Sub(now) <= rule.LeadTime {
		err := service.trainingRepo.CreateFromOccurrence(ctx, rule.Id, rule.Description, occurrence, rule.MaxQuantity, rule.Duration)
		if err != nil {
			return err
		}
		occurrence = occurrence.Add(week)
	}

	return nil
}
