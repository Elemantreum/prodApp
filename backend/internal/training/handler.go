package training

import (
	"encoding/json"
	"net/http"
	"prodApp/internal/platform/apperrors"
	http2 "prodApp/internal/platform/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type TrainingHandler struct {
	TrainingService *TrainingService
}

type createTrainingRequest struct {
	Description string        `json:"description"`
	Date        time.Time     `json:"date" format:"date-time" example:"2002-10-02T15:00:00Z"`
	MaxQuantity int           `json:"maxQuantity"`
	Duration    time.Duration `json:"duration" swaggertype:"primitive,integer"`
}

type updateTrainingRequest struct {
	Id          int           `json:"id"`
	Description string        `json:"description"`
	Date        time.Time     `json:"date" format:"date-time" example:"2002-10-02T15:00:00Z"`
	MaxQuantity int           `json:"maxQuantity"`
	Duration    time.Duration `json:"duration" swaggertype:"primitive,integer"`
	IsActive    bool          `json:"isActive"`
}

type trainingIdRequest struct {
	Id int `json:"id"`
}

type createRecurringTrainingRequest struct {
	Description string        `json:"description"`
	AnchorDate  time.Time     `json:"anchorDate" format:"date-time" example:"2002-10-02T15:00:00Z"`
	MaxQuantity int           `json:"maxQuantity"`
	Duration    time.Duration `json:"duration" swaggertype:"primitive,integer"`
	LeadTime    time.Duration `json:"leadTime" swaggertype:"primitive,integer"`
}

type updateRecurringTrainingRequest struct {
	Id          int           `json:"id"`
	Description string        `json:"description"`
	AnchorDate  time.Time     `json:"anchorDate" format:"date-time" example:"2002-10-02T15:00:00Z"`
	MaxQuantity int           `json:"maxQuantity"`
	Duration    time.Duration `json:"duration" swaggertype:"primitive,integer"`
	LeadTime    time.Duration `json:"leadTime" swaggertype:"primitive,integer"`
	IsActive    bool          `json:"isActive"`
}

func NewTrainingHandler(trainingService *TrainingService) *TrainingHandler {
	return &TrainingHandler{TrainingService: trainingService}
}

// CreateTraining godoc
// @Summary      Создать тренировку
// @Description  Только для админов
// @Tags         trainings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Param        request  body      createTrainingRequest  true  "Данные тренировки"
// @Success      200      {object}  Training
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "нет прав администратора"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /trainings/create [post]
func (h *TrainingHandler) CreateTraining(w http.ResponseWriter, r *http.Request) {
	var req createTrainingRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "createTrainingRequest"})
		return
	}

	saved, err := h.TrainingService.AddTraining(r.Context(), req.Description, req.Date, req.MaxQuantity, req.Duration)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, saved)
}

// UpdateTraining godoc
// @Summary      Изменить тренировку
// @Description  Только для админов
// @Tags         trainings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Param        request  body      updateTrainingRequest  true  "Изменённые данные тренировки"
// @Success      200      {object}  Training
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "нет прав администратора"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /trainings/update [post]
func (h *TrainingHandler) UpdateTraining(w http.ResponseWriter, r *http.Request) {
	var req updateTrainingRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "updateTrainingRequest"})
		return
	}

	updated, err := h.TrainingService.UpdateTraining(r.Context(), req.Id, req.Description, req.Date, req.MaxQuantity, req.Duration, req.IsActive)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, updated)
}

// DeleteTraining godoc
// @Summary      Удалить тренировку
// @Description  Только для админов. Отказывает (409), если на тренировку есть хоть одна запись — сначала деактивируйте её через UpdateTraining.
// @Tags         trainings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Param        request  body      trainingIdRequest  true  "ID тренировки"
// @Success      200      {object}  bool
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "нет прав администратора"
// @Failure      409      {object}  map[string]string  "на тренировку есть записи"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /trainings/delete [post]
func (h *TrainingHandler) DeleteTraining(w http.ResponseWriter, r *http.Request) {
	var req trainingIdRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "trainingIdRequest"})
		return
	}

	err = h.TrainingService.DeleteTraining(r.Context(), req.Id)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, true)
}

// CreateRecurringTraining godoc
// @Summary      Создать регулярную тренировку
// @Description  Только для админов. AnchorDate — дата и время первого занятия, дальше повторяется каждые 7 дней от неё. Первые занятия материализуются сразу.
// @Tags         trainings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Param        request  body      createRecurringTrainingRequest  true  "Данные правила"
// @Success      200      {object}  RecurringTraining
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "нет прав администратора"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /trainings/recurring/create [post]
func (h *TrainingHandler) CreateRecurringTraining(w http.ResponseWriter, r *http.Request) {
	var req createRecurringTrainingRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "createRecurringTrainingRequest"})
		return
	}

	saved, err := h.TrainingService.CreateRule(r.Context(), req.Description, req.AnchorDate, req.MaxQuantity, req.Duration, req.LeadTime)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, saved)
}

// GetRecurringTrainings godoc
// @Summary      Список правил регулярных тренировок
// @Description  Только для админов
// @Tags         trainings
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Success      200  {array}   RecurringTraining
// @Failure      401  {object}  map[string]string  "не авторизован"
// @Failure      403  {object}  map[string]string  "нет прав администратора"
// @Failure      500  {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /trainings/recurring/get [get]
func (h *TrainingHandler) GetRecurringTrainings(w http.ResponseWriter, r *http.Request) {
	rules, err := h.TrainingService.ListRules(r.Context())

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, rules)
}

// UpdateRecurringTraining godoc
// @Summary      Изменить регулярную тренировку
// @Description  Только для админов. Тем же полем isActive правило можно как остановить, так и снова включить.
// @Tags         trainings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Param        request  body      updateRecurringTrainingRequest  true  "Изменённые данные правила"
// @Success      200      {object}  RecurringTraining
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "нет прав администратора"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /trainings/recurring/update [post]
func (h *TrainingHandler) UpdateRecurringTraining(w http.ResponseWriter, r *http.Request) {
	var req updateRecurringTrainingRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "updateRecurringTrainingRequest"})
		return
	}

	updated, err := h.TrainingService.UpdateRule(r.Context(), req.Id, req.Description, req.AnchorDate, req.MaxQuantity, req.Duration, req.LeadTime, req.IsActive)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, updated)
}

// GetAllTrainings godoc
// @Summary      Получить список тренировок
// @Description  Админу — все тренировки, включая неактивные. Остальным авторизованным пользователям — только активные.
// @Tags         trainings
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Success      200  {array}   Training
// @Failure      401  {object}  map[string]string  "не авторизован"
// @Failure      500  {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /trainings/get [get]
func (h *TrainingHandler) GetAllTrainings(w http.ResponseWriter, r *http.Request) {
	claims, _ := http2.ClaimsFromContext(r.Context())

	trainings, err := h.TrainingService.GetAll(r.Context(), claims)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, trainings)
}

// GetTrainingById godoc
// @Summary      Получить тренировку по ID
// @Description  Только для админов
// @Tags         trainings
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Param        id   path      int  true  "ID тренировки"
// @Success      200  {object}  Training
// @Failure      401  {object}  map[string]string  "не авторизован"
// @Failure      403  {object}  map[string]string  "нет прав администратора"
// @Failure      500  {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /trainings/get/{id} [get]
func (h *TrainingHandler) GetTrainingById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)

	trainings, err := h.TrainingService.FindById(r.Context(), id)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, trainings)
}
