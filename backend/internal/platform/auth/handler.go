package auth

import (
	"encoding/json"
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
	userService *user.UserService
}

func NewAuthHandler(tgAuth *TelegramAuthenticator, vkAuth *VKAuthenticator, tokenIssuer *TokenIssuer, userService *user.UserService) *AuthHandler {
	return &AuthHandler{
		tgAuth:      tgAuth,
		vkAuth:      vkAuth,
		tokenIssuer: tokenIssuer,
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

	appUser, err := h.userService.ResolveAndLink(r.Context(), code, principal.Provider, principal.ExternalId)
	if err != nil {
		http2.RespondError(w, err)
		return
	}

	token, err := h.tokenIssuer.IssueAccessToken(appUser.Id, appUser.IsAdmin, appUser.IsApproved)
	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, tokenResponse{AccessToken: token})
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *AuthHandler) issueAndRespond(w http.ResponseWriter, r *http.Request, principal *Principal) {
	appUser, err := h.userService.EnsureProfile(r.Context(), principal.Provider, principal.ExternalId, principal.FirstName, principal.LastName)
	if err != nil {
		http2.RespondError(w, err)
		return
	}

	token, err := h.tokenIssuer.IssueAccessToken(appUser.Id, appUser.IsAdmin, appUser.IsApproved)
	if err != nil {
		http2.RespondError(w, err)
		return
	}

	http2.RespondJSON(w, http.StatusOK, tokenResponse{AccessToken: token})
}
