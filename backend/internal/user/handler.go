package user

import (
	"encoding/json"
	"net/http"
	"prodApp/internal/platform/apperrors"
	http2 "prodApp/internal/platform/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	UserService *UserService
}

type userIdRequest struct {
	UserId int `json:"userId"`
}

type profileRequest struct {
	UserId    int    `json:"userId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type identityInfo struct {
	Provider   string `json:"provider"`
	ExternalId string `json:"externalId"`
	Username   string `json:"username"`
}

type userWithIdentities struct {
	User
	Identities []identityInfo `json:"identities"`
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{UserService: service}
}

type linkCodeResponse struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// GenerateLinkCode godoc
// @Summary      Получить код для привязки другой платформы
// @Description  Доступно любому авторизованному пользователю. Код одноразовый, живёт 5 минут — ввести его нужно при входе через ДРУГУЮ платформу (Telegram/VK) в /auth/telegram/link или /auth/vk/link, чтобы объединить аккаунты.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  linkCodeResponse
// @Failure      401  {object}  map[string]string  "не авторизован"
// @Failure      500  {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /user/linkCode [post]
func (h *UserHandler) GenerateLinkCode(w http.ResponseWriter, r *http.Request) {
	claims, ok := http2.ClaimsFromContext(r.Context())

	if !ok {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	code, err := h.UserService.GenerateLinkCode(r.Context(), claims.UserID)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, linkCodeResponse{Code: code.Code, ExpiresAt: code.ExpiresAt})
}

// GetOwnProfile godoc
// @Summary      Мой профиль
// @Description  Доступно любому авторизованному пользователю, только про самого себя
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  userWithIdentities
// @Failure      401  {object}  map[string]string  "не авторизован"
// @Failure      500  {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /user/me [get]
func (h *UserHandler) GetOwnProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := http2.ClaimsFromContext(r.Context())

	if !ok {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	u, err := h.UserService.GetUser(r.Context(), claims.UserID)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	identities, err := h.UserService.GetIdentities(r.Context(), u.Id)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	infos := make([]identityInfo, 0, len(identities))
	for _, id := range identities {
		infos = append(infos, identityInfo{Provider: id.Provider, ExternalId: id.ExternalId, Username: id.Username})
	}

	http2.RespondJSON(w, http.StatusOK, userWithIdentities{User: *u, Identities: infos})
}

// GetOwnIdentities godoc
// @Summary      Какие платформы привязаны к моему аккаунту
// @Description  Доступно любому авторизованному пользователю, только про самого себя
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   identityInfo
// @Failure      401  {object}  map[string]string  "не авторизован"
// @Failure      500  {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /user/identities [get]
func (h *UserHandler) GetOwnIdentities(w http.ResponseWriter, r *http.Request) {
	claims, ok := http2.ClaimsFromContext(r.Context())

	if !ok {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	identities, err := h.UserService.GetIdentities(r.Context(), claims.UserID)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	infos := make([]identityInfo, 0, len(identities))
	for _, id := range identities {
		infos = append(infos, identityInfo{Provider: id.Provider, ExternalId: id.ExternalId, Username: id.Username})
	}

	http2.RespondJSON(w, http.StatusOK, infos)
}

// UpdateProfile godoc
// @Summary      Изменить имя и фамилию
// @Description  userId в теле — 0 или отсутствует значит "себя". Обычный пользователь может менять только себя, и только пока не одобрен администратором — после одобрения имя "замораживается" и дальше его может поменять только админ (кому угодно и когда угодно).
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      profileRequest  true  "ID пользователя (необязательно), имя и фамилия"
// @Success      200      {object}  userWithIdentities
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "чужой профиль или уже одобрен"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /user/updateProfile [post]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req profileRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "userId/firstName/lastName"})
		return
	}

	claims, ok := http2.ClaimsFromContext(r.Context())

	if !ok {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	targetId := req.UserId
	if targetId == 0 {
		targetId = claims.UserID
	}

	updated, err := h.UserService.UpdateProfile(r.Context(), targetId, req.FirstName, req.LastName, claims)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	identities, err := h.UserService.GetIdentities(r.Context(), updated.Id)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	infos := make([]identityInfo, 0, len(identities))
	for _, id := range identities {
		infos = append(infos, identityInfo{Provider: id.Provider, ExternalId: id.ExternalId, Username: id.Username})
	}

	http2.RespondJSON(w, http.StatusOK, userWithIdentities{User: *updated, Identities: infos})
}

// GetAllUsers godoc
// @Summary      Получить список пользователей
// @Description  Только для админов
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   userWithIdentities
// @Failure      401  {object}  map[string]string  "не авторизован"
// @Failure      403  {object}  map[string]string  "нет прав администратора"
// @Failure      500  {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /user/get [get]
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserService.ListUsers(r.Context())

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	result := make([]userWithIdentities, 0, len(users))

	for _, u := range users {
		identities, err := h.UserService.GetIdentities(r.Context(), u.Id)

		if err != nil {
			http2.RespondError(w, err)
			return
		}

		infos := make([]identityInfo, 0, len(identities))
		for _, id := range identities {
			infos = append(infos, identityInfo{Provider: id.Provider, ExternalId: id.ExternalId, Username: id.Username})
		}

		result = append(result, userWithIdentities{User: u, Identities: infos})
	}

	http2.RespondJSON(w, http.StatusOK, result)
}

// GetUserById godoc
// @Summary      Получить пользователя по ID
// @Description  Только для админов
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID пользователя"
// @Success      200  {object}  userWithIdentities
// @Failure      401  {object}  map[string]string  "не авторизован"
// @Failure      403  {object}  map[string]string  "нет прав администратора"
// @Failure      404  {object}  map[string]string  "пользователь не найден"
// @Failure      500  {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /user/get/{id} [get]
func (h *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "id"})
		return
	}

	u, err := h.UserService.GetUser(r.Context(), id)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	identities, err := h.UserService.GetIdentities(r.Context(), u.Id)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	infos := make([]identityInfo, 0, len(identities))
	for _, id := range identities {
		infos = append(infos, identityInfo{Provider: id.Provider, ExternalId: id.ExternalId, Username: id.Username})
	}

	http2.RespondJSON(w, http.StatusOK, userWithIdentities{User: *u, Identities: infos})
}

// ApproveUser godoc
// @Summary      Одобрить пользователя
// @Description  Только для админов
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      userIdRequest  true  "ID пользователя"
// @Success      200      {object}  userWithIdentities
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "нет прав администратора"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /user/approve [post]
func (h *UserHandler) ApproveUser(w http.ResponseWriter, r *http.Request) {
	var req userIdRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "userId"})
		return
	}

	err = h.UserService.ApproveUser(r.Context(), req.UserId)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	updated, err := h.UserService.GetUser(r.Context(), req.UserId)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	identities, err := h.UserService.GetIdentities(r.Context(), updated.Id)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	infos := make([]identityInfo, 0, len(identities))
	for _, id := range identities {
		infos = append(infos, identityInfo{Provider: id.Provider, ExternalId: id.ExternalId, Username: id.Username})
	}

	http2.RespondJSON(w, http.StatusOK, userWithIdentities{User: *updated, Identities: infos})
}

// UnapproveUser godoc
// @Summary      Снять одобрение с пользователя
// @Description  Только для админов
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      userIdRequest  true  "ID пользователя"
// @Success      200      {object}  userWithIdentities
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "нет прав администратора"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /user/unapprove [post]
func (h *UserHandler) UnapproveUser(w http.ResponseWriter, r *http.Request) {
	var req userIdRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "userId"})
		return
	}

	err = h.UserService.UnapproveUser(r.Context(), req.UserId)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	updated, err := h.UserService.GetUser(r.Context(), req.UserId)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	identities, err := h.UserService.GetIdentities(r.Context(), updated.Id)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	infos := make([]identityInfo, 0, len(identities))
	for _, id := range identities {
		infos = append(infos, identityInfo{Provider: id.Provider, ExternalId: id.ExternalId, Username: id.Username})
	}

	http2.RespondJSON(w, http.StatusOK, userWithIdentities{User: *updated, Identities: infos})
}

// MakeAdmin godoc
// @Summary      Выдать права администратора
// @Description  Только для админов
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      userIdRequest  true  "ID пользователя"
// @Success      200      {object}  userWithIdentities
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "нет прав администратора"
// @Failure      409      {object}  map[string]string  "уже администратор"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /user/makeAdmin [post]
func (h *UserHandler) MakeAdmin(w http.ResponseWriter, r *http.Request) {
	var req userIdRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "userId"})
		return
	}

	err = h.UserService.MakeAdmin(r.Context(), req.UserId)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	updated, err := h.UserService.GetUser(r.Context(), req.UserId)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	identities, err := h.UserService.GetIdentities(r.Context(), updated.Id)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	infos := make([]identityInfo, 0, len(identities))
	for _, id := range identities {
		infos = append(infos, identityInfo{Provider: id.Provider, ExternalId: id.ExternalId, Username: id.Username})
	}

	http2.RespondJSON(w, http.StatusOK, userWithIdentities{User: *updated, Identities: infos})
}

// RemoveAdmin godoc
// @Summary      Забрать права администратора
// @Description  Только для админов. Нельзя применить к самому себе.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      userIdRequest  true  "ID пользователя"
// @Success      200      {object}  userWithIdentities
// @Failure      400      {object}  map[string]string  "нельзя снять права с самого себя"
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "нет прав администратора"
// @Failure      409      {object}  map[string]string  "и так не администратор"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /user/removeAdmin [post]
func (h *UserHandler) RemoveAdmin(w http.ResponseWriter, r *http.Request) {
	var req userIdRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "userId"})
		return
	}

	claims, ok := http2.ClaimsFromContext(r.Context())

	if !ok {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	if claims.UserID == req.UserId {
		http2.RespondError(w, &apperrors.ValidationError{Field: "userId: нельзя применить это действие к самому себе"})
		return
	}

	err = h.UserService.RemoveAdmin(r.Context(), req.UserId)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	updated, err := h.UserService.GetUser(r.Context(), req.UserId)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	identities, err := h.UserService.GetIdentities(r.Context(), updated.Id)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	infos := make([]identityInfo, 0, len(identities))
	for _, id := range identities {
		infos = append(infos, identityInfo{Provider: id.Provider, ExternalId: id.ExternalId, Username: id.Username})
	}

	http2.RespondJSON(w, http.StatusOK, userWithIdentities{User: *updated, Identities: infos})
}

// DeleteUser godoc
// @Summary      Удалить пользователя
// @Description  Только для админов. Нельзя применить к самому себе. Каскадно удаляет его брони.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      userIdRequest  true  "ID пользователя"
// @Success      200      {boolean} true
// @Failure      400      {object}  map[string]string  "нельзя удалить самого себя"
// @Failure      401      {object}  map[string]string  "не авторизован"
// @Failure      403      {object}  map[string]string  "нет прав администратора"
// @Failure      500      {object}  map[string]string  "внутренняя ошибка сервера"
// @Router       /user/delete [post]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var req userIdRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "userId"})
		return
	}

	claims, ok := http2.ClaimsFromContext(r.Context())

	if !ok {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	if claims.UserID == req.UserId {
		http2.RespondError(w, &apperrors.ValidationError{Field: "userId: нельзя применить это действие к самому себе"})
		return
	}

	err = h.UserService.DeleteUser(r.Context(), req.UserId)

	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, true)
}
