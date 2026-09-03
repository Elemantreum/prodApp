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

// GetAllTrainings godoc
// @Summary      Получить список тренировок
// @Description  Только для админов
// @Tags         trainings
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Success      200  {array}   Training
// @Failure      401  {object}  map[string]string  "не авторизован"
// @Failure      403  {object}  map[string]string  "нет прав администратора"
// @Failure      500  {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /trainings/get [get]
func (h *TrainingHandler) GetAllTrainings(w http.ResponseWriter, r *http.Request) {

	trainings, err := h.TrainingService.GetAll(r.Context())

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
