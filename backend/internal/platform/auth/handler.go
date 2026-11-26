package auth

import (
	"encoding/json"
	"net/http"
	"prodApp/internal/platform/apperrors"
	"strings"
	"time"

	http2 "prodApp/internal/platform/http"
)

type AuthHandler struct {
	tgAuth  *TelegramAuthenticator
	vkAuth  *VKAuthenticator
	service *AuthService
	botName string
}

func NewAuthHandler(tgAuth *TelegramAuthenticator, vkAuth *VKAuthenticator, service *AuthService, botName string) *AuthHandler {
	return &AuthHandler{tgAuth: tgAuth, vkAuth: vkAuth, service: service, botName: botName}
}

func (h *AuthHandler) WebsiteBot(w http.ResponseWriter, _ *http.Request) {
	http2.RespondJSON(w, http.StatusOK, map[string]string{"bot": h.botName})
}

// TelegramLogin godoc
// @Summary      Логин через Telegram Mini App
// @Tags         auth
// @Produce      json
// @Success      200  {object}  tokenResponse
// @Failure      401
// @Router       /auth/telegram [post]
func (h *AuthHandler) TelegramLogin(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	initData := strings.TrimPrefix(authHeader, "tma ")

	principal, err := h.tgAuth.Verify(initData)
	if err != nil {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	tokens, err := h.service.Login(r.Context(), principal)
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, tokens)
}

// VKLogin godoc
// @Summary      Логин через VK Mini App
// @Tags         auth
// @Produce      json
// @Success      200  {object}  tokenResponse
// @Failure      401
// @Router       /auth/vk [post]
func (h *AuthHandler) VKLogin(w http.ResponseWriter, r *http.Request) {
	principal, err := h.vkAuth.Verify(r.URL.RawQuery)
	if err != nil {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	tokens, err := h.service.Login(r.Context(), principal)
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, tokens)
}

// TelegramLink godoc
// @Summary      Привязать Telegram к существующему аккаунту
// @Description  Код берётся из POST /user/linkCode, полученного в сессии ДРУГОЙ платформы. После успешной привязки возвращает токен для аккаунта, к которому привязали (не нового).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      linkRequest  true  "Код привязки"
// @Success      200      {object}  tokenResponse
// @Failure      400      {object}  map[string]string  "код не указан или истёк"
// @Failure      401      {object}  map[string]string  "невалидные initData"
// @Router       /auth/telegram/link [post]
func (h *AuthHandler) TelegramLink(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	initData := strings.TrimPrefix(authHeader, "tma ")

	principal, err := h.tgAuth.Verify(initData)
	if err != nil {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	var req linkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "code"})
		return
	}

	tokens, err := h.service.Link(r.Context(), principal, req.Code)
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, tokens)
}

// VKLink godoc
// @Summary      Привязать VK к существующему аккаунту
// @Description  Код берётся из POST /user/linkCode, полученного в сессии ДРУГОЙ платформы, и передаётся как query-параметр code вместе с обычными launch-параметрами VK. После успешной привязки возвращает токен для аккаунта, к которому привязали (не нового).
// @Tags         auth
// @Produce      json
// @Param        code  query     string  true  "Код привязки"
// @Success      200   {object}  tokenResponse
// @Failure      400   {object}  map[string]string  "код не указан или истёк"
// @Failure      401   {object}  map[string]string  "невалидные launch params"
// @Router       /auth/vk/link [post]
func (h *AuthHandler) VKLink(w http.ResponseWriter, r *http.Request) {
	principal, err := h.vkAuth.Verify(r.URL.RawQuery)
	if err != nil {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}

	code := r.URL.Query().Get("code")

	tokens, err := h.service.Link(r.Context(), principal, code)
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, tokens)
}

type linkRequest struct {
	Code string `json:"code"`
}

// Refresh godoc
// @Summary      Обновить access token по refresh token'у
// @Description  Ротация: предъявленный refresh token сразу отзывается, взамен выдаётся новая пара. Повторное предъявление того же токена (например, украденного) больше не сработает.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      refreshRequest  true  "Refresh token"
// @Success      200      {object}  tokenResponse
// @Failure      400      {object}  map[string]string  "refresh_token не указан"
// @Failure      401      {object}  map[string]string  "невалидный/истёкший/отозванный refresh token"
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		http2.RespondError(w, &apperrors.ValidationError{Field: "refresh_token"})
		return
	}

	tokens, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, tokens)
}

// Logout godoc
// @Summary      Отозвать refresh token
// @Tags         auth
// @Accept       json
// @Success      204
// @Failure      400  {object}  map[string]string  "refresh_token не указан"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		http2.RespondError(w, &apperrors.ValidationError{Field: "refresh_token"})
		return
	}

	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		http2.RespondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) WebsiteLogin(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	if err := r.ParseForm(); err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "telegram"})
		return
	}
	principal, err := h.tgAuth.websitePrincipal(r.PostForm, time.Now())
	if err != nil {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}
	tokens, err := h.service.Login(r.Context(), principal)
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, tokens)
}
