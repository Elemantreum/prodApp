package booking

import (
	"encoding/json"
	"net/http"
	"prodApp/internal/platform/apperrors"
	http2 "prodApp/internal/platform/http"
	"strconv"
)

type BookingHandler struct {
	BookingService *BookingService
}

type signUpForTrainingRequest struct {
	TrainingId int `json:"trainingId"`
}
type cancelBookingRequest struct {
	BookingId int `json:"bookingId"`
}
type setAttendanceRequest struct {
	BookingId int  `json:"bookingId"`
	Attended  bool `json:"attended"`
}

func NewBookingHandler(service *BookingService) *BookingHandler {
	return &BookingHandler{BookingService: service}
}

// SignupForTraining godoc
// @Summary      Записаться на тренировку
// @Description  Создаёт бронирование текущего пользователя на тренировку
// @Tags         booking
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Param        request  body      signUpForTrainingRequest  true  "ID тренировки"
// @Success      200      {object}  Booking
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "аккаунт ещё не одобрен администратором"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /booking/signUp [post]
func (handler *BookingHandler) SignupForTraining(w http.ResponseWriter, r *http.Request) {
	var req signUpForTrainingRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "signUpForTrainingRequest"})
	}

	claims, ok := http2.ClaimsFromContext(r.Context())

	if !ok {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	saved, err := handler.BookingService.CreateBooking(r.Context(), claims.UserID, req.TrainingId)

	if err != nil {
		http2.RespondError(w, err)
	}

	http2.RespondJSON(w, http.StatusOK, saved)
}

// GetBookingsByUserId godoc
// @Summary      Получить бронирования пользователя
// @Description  Доступно администратору для любого пользователя; обычному пользователю — только для самого себя
// @Tags         booking
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Param        userId  query     int  true  "ID пользователя"
// @Success      200     {array}   Booking
// @Failure      401     {object}  map[string]string  "не авторизован"
// @Failure      403     {object}  map[string]string  "доступ запрещён"
// @Failure      500     {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /booking/getByUserId [get]
func (handler *BookingHandler) GetBookingsByUserId(w http.ResponseWriter, r *http.Request) {
	claims, ok := http2.ClaimsFromContext(r.Context())

	if !ok {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	userId, err := strconv.Atoi(r.URL.Query().Get("userId"))
	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "userId"})
		return
	}

	bookings, err := handler.BookingService.GetAllBookingsByUserId(r.Context(), userId, claims)

	if err != nil {
		http2.RespondError(w, err)
	}

	http2.RespondJSON(w, http.StatusOK, bookings)
}

// GetParticipants godoc
// @Summary      Кто записан на тренировку
// @Description  Доступно любому авторизованному пользователю. Показывает только активные записи.
// @Tags         booking
// @Produce      json
// @Security     BearerAuth
// @Param        trainingId  query     int  true  "ID тренировки"
// @Success      200         {array}   Participant
// @Failure      401         {object}  map[string]string  "не авторизован"
// @Failure      500         {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /booking/getByTrainingId [get]
func (handler *BookingHandler) GetParticipants(w http.ResponseWriter, r *http.Request) {
	trainingId, err := strconv.Atoi(r.URL.Query().Get("trainingId"))

	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "trainingId"})
		return
	}

	participants, err := handler.BookingService.GetParticipants(r.Context(), trainingId)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, participants)
}

// GetTrainingBookings godoc
// @Summary      Все записи на тренировку (админ)
// @Description  Только для админов. В отличие от getByTrainingId — включает отменённые записи и id брони.
// @Tags         booking
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Param        trainingId  query     int  true  "ID тренировки"
// @Success      200         {array}   BookingDetail
// @Failure      401         {object}  map[string]string  "не авторизован"
// @Failure      403         {object}  map[string]string  "нет прав администратора"
// @Failure      500         {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /booking/admin/getByTrainingId [get]
func (handler *BookingHandler) GetTrainingBookings(w http.ResponseWriter, r *http.Request) {
	trainingId, err := strconv.Atoi(r.URL.Query().Get("trainingId"))

	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "trainingId"})
		return
	}

	bookings, err := handler.BookingService.GetTrainingBookings(r.Context(), trainingId)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, bookings)
}

// SetAttendance godoc
// @Summary      Отметить посещаемость (админ)
// @Description  Только для админов. Переводит бронь в completed (пришёл) или not_completed (не пришёл); отменённую бронь трогать нельзя.
// @Tags         booking
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Param        request  body      setAttendanceRequest  true  "ID брони и признак посещения"
// @Success      200      {object}  Booking
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "нет прав администратора"
// @Failure      409      {object}  map[string]string  "бронь отменена"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /booking/setAttendance [post]
func (handler *BookingHandler) SetAttendance(w http.ResponseWriter, r *http.Request) {
	var req setAttendanceRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "setAttendanceRequest"})
		return
	}

	updated, err := handler.BookingService.SetAttendance(r.Context(), req.BookingId, req.Attended)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, updated)
}

// CancelBooking godoc
// @Summary      Отменить бронирование
// @Description  Доступно администратору для любого бронирования; обычному пользователю — только для своего
// @Tags         booking
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Security     OAuth2AccessCode
// @Param        request  body      cancelBookingRequest  true  "ID бронирования"
// @Success      200      {object}  Booking
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "доступ запрещён"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /booking/cancel [post]
func (handler *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	var req cancelBookingRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "cancelBookingRequest"})
	}

	claims, ok := http2.ClaimsFromContext(r.Context())

	if !ok {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	err = handler.BookingService.CancelBooking(r.Context(), req.BookingId, claims)

	if err != nil {
		http2.RespondError(w, err)
	}

	http2.RespondJSON(w, http.StatusOK, true)
}
