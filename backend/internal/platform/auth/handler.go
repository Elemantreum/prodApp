package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"prodApp/internal/platform/apperrors"
	"strings"

	http2 "prodApp/internal/platform/http"
	"prodApp/internal/user"
)

type AuthHandler struct {
	tgAuth      *TelegramAuthenticator
	vkAuth      *VKAuthenticator
	tokenIssuer *TokenIssuer
	refreshRepo *RefreshTokenRepository
	userService *user.UserService
}

func NewAuthHandler(tgAuth *TelegramAuthenticator, vkAuth *VKAuthenticator, tokenIssuer *TokenIssuer, refreshRepo *RefreshTokenRepository, userService *user.UserService) *AuthHandler {
	return &AuthHandler{
		tgAuth:      tgAuth,
		vkAuth:      vkAuth,
		tokenIssuer: tokenIssuer,
		refreshRepo: refreshRepo,
		userService: userService,
	}
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

	h.issueAndRespond(w, r, principal)
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

	h.issueAndRespond(w, r, principal)
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

	h.linkAndRespond(w, r, principal, req.Code)
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

	h.linkAndRespond(w, r, principal, code)
}

type linkRequest struct {
	Code string `json:"code"`
}

func (h *AuthHandler) linkAndRespond(w http.ResponseWriter, r *http.Request, principal *Principal, code string) {
	if code == "" {
		http2.RespondError(w, &apperrors.ValidationError{Field: "code"})
		return
	}

	appUser, err := h.userService.ResolveAndLink(r.Context(), code, principal.Provider, principal.ExternalId, principal.Username)
	if err != nil {
		http2.RespondError(w, err)
		return
	}

	h.issueTokenPairAndRespond(w, r, appUser)
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) issueAndRespond(w http.ResponseWriter, r *http.Request, principal *Principal) {
	appUser, err := h.userService.EnsureProfile(r.Context(), principal.Provider, principal.ExternalId, principal.FirstName, principal.LastName, principal.Username)
	if err != nil {
		http2.RespondError(w, err)
		return
	}

	h.issueTokenPairAndRespond(w, r, appUser)
}

func (h *AuthHandler) issueTokenPairAndRespond(w http.ResponseWriter, r *http.Request, appUser *user.User) {
	access, err := h.tokenIssuer.IssueAccessToken(appUser.Id, appUser.IsAdmin, appUser.IsApproved)
	if err != nil {
		http2.RespondError(w, err)
		return
	}

	refresh, err := h.refreshRepo.Create(r.Context(), appUser.Id, h.tokenIssuer.RefreshTokenTTL())
	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, tokenResponse{AccessToken: access, RefreshToken: refresh})
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

	userID, err := h.refreshRepo.FindActiveUserID(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenInvalid) {
			http2.RespondError(w, &apperrors.UnauthorizedError{})
			return
		}
		http2.RespondError(w, err)
		return
	}

	if err := h.refreshRepo.Revoke(r.Context(), req.RefreshToken); err != nil {
		http2.RespondError(w, err)
		return
	}

	appUser, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		http2.RespondError(w, err)
		return
	}

	h.issueTokenPairAndRespond(w, r, appUser)
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

	if err := h.refreshRepo.Revoke(r.Context(), req.RefreshToken); err != nil {
		http2.RespondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
