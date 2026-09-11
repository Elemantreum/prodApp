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
	// CreateFromOccurrence — nil, nil (не ошибка), если строка уже
	// существует (ON CONFLICT DO NOTHING поверх UNIQUE(recurring_training_id,
	// date) — идемпотентность повторных проходов GenerateOccurrences, см.
	// generateForRule). Именно поэтому не просто error — вызывающему нужно
	// отличить "ничего не создали" от "создали", чтобы не публиковать
	// OnCreateTraining на пустом месте.
	CreateFromOccurrence(ctx context.Context, recurringId int, description string, date time.Time, maxQuantity int, duration time.Duration) (*Training, error)
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

// ChatIdResolver — узкий порт наружу в user-пакет (реализован
// *user.PostgresIdentityRepository — структурная типизация, отдельного
// импорта user отсюда не требуется). Нужен ровно один метод: "дай telegram
// chat id всех, кому интересны тренировки" — результат кладётся прямо в
// payload события (OnCreateTraining/OnUpdateTraining), чтобы tg-notifier
// мог разослать уведомления без собственного доступа к Postgres.
type ChatIdResolver interface {
	AllTelegramChatIds(ctx context.Context) ([]int64, error)
}

// TrainingService — CRUD обычных тренировок И управление регулярными
// правилами разом (было двумя сервисами, RecurringService слит сюда):
// генерация occurrences — это просто ещё один способ создать Training,
// который должен проходить через тот же путь публикации OnCreateTraining,
// что и AddTraining, — раньше это было не так (generateForRule дёргал
// репозиторий напрямую), из-за чего уведомление не приходило на
// автоматически сгенерированные регулярные занятия. Один сервис — один
// путь создания (createAndPublish), в котором эта логика гарантированно
// не разъедется снова.
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

// publishCreateEvent — единственное место, откуда уходит OnCreateTraining:
// и обычное ручное создание (AddTraining), и автогенерация occurrences
// регулярного правила (generateForRule) идут через него, поэтому
// уведомление гарантированно приходит на ЛЮБОЕ создание тренировки, а не
// только на ручное.
func (s *TrainingService) publishCreateEvent(ctx context.Context, t *Training) {
	chatIds, err := s.chatIds.AllTelegramChatIds(ctx)
	if err != nil {
		// Не фейлим создание тренировки из-за этого — просто разошлём
		// событие без адресатов, tg-notifier никого не уведомит в этот раз.
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

	// Сбой записи в outbox не должен проваливать уже состоявшееся создание
	// тренировки — тренировка реально создана, клиенту это нужно увидеть как
	// успех. Пока Create не в одной транзакции с repo.Create (см. пакет
	// outbox), возможна нестыковка "тренировка есть, события нет" — это
	// осознанно оставлено на потом.
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

	// Снимок ДО изменения — нужен, чтобы событие несло diff, а не просто
	// новое состояние (см. OnUpdateTraining).
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

// GetAll возвращает все тренировки админу и только активные — всем
// остальным. requester может быть nil (на всякий случай, если middleware
// почему-то не положил claims в контекст) — тогда, как и для любого
// обычного пользователя, неактивные тренировки не отдаются.
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

// DeleteTraining удаляет тренировку насовсем. Отказывает, если на неё есть
// хоть одна бронь (активная, отменённая, отмеченная посещаемостью — любая):
// это уже история, каскадное удаление через FK молча снесло бы её вместе с
// тренировкой. Деактивация (UpdateTraining с isActive=false) — правильный
// способ "убрать" тренировку с сохранением истории; Delete — только для
// пустых, ничего не значащих записей (опечатка, лишнее регулярное занятие
// и т.п.).
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

// DeactivatePastTrainings — вызывается периодически фоновым тикером (см.
// main.go): переводит прошедшие тренировки в isActive=false, чтобы они сами
// переставали быть видны обычным пользователям (см. GetAll), без ручного
// вмешательства админа. Booking.CreateBooking всё равно дополнительно
// проверяет дату напрямую — на случай гонки в окне между тем, как
// тренировка прошла, и следующим тиком этого метода.
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

// UpdateRule меняет все поля правила разом, включая isActive — так
// включение/выключение регулярности и правка остальных полей (даты первого
// занятия, мест, длительности, времени создания) идут через одну и ту же
// ручку, а не отдельный toggle. Уже созданные тренировки не трогает — они
// остаются обычными строками в trainings, которые админ может отдельно
// отредактировать/деактивировать через обычный UpdateTraining. Если правило
// снова активно и изменения открыли новые "созревшие" occurrences (например
// сдвинули anchorDate назад или увеличили leadTime), они материализуются
// сразу же, не дожидаясь фонового тика.
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

// GenerateOccurrences — идемпотентный проход по всем активным правилам:
// досоздаёт недостающие занятия, чьё окно создания уже наступило (см.
// generateForRule и поле LeadTime). Вызывается и по расписанию (см.
// main.go), и сразу при создании/изменении правила. Безопасно звать сколько
// угодно раз — INSERT ... ON CONFLICT DO NOTHING на стороне репозитория
// гарантирует отсутствие дублей.
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

// generateForRule материализует все occurrences правила, чьё "окно
// создания" уже открылось — то есть occurrence.Sub(now) <= rule.LeadTime.
// Occurrences идут по возрастанию с фиксированным недельным шагом, так что
// как только встретили ещё не созревший — все следующие тем более не
// созрели, можно останавливаться. Отстающих (например после простоя
// сервера) это тоже само подхватывает: цикл просто стартует от anchor и
// создаёт всё, что уже пора, сколько бы occurrences ни накопилось.
//
// Каждая реально созданная (неno-op, см. CreateFromOccurrence) occurrence
// идёт через тот же publishCreateEvent, что и ручное AddTraining — именно
// это чинит "уведомление не приходит на регулярные тренировки", ради чего
// RecurringService и слит в TrainingService.
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
