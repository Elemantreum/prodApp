package sectioninvite

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"prodApp/internal/platform/apperrors"
	http2 "prodApp/internal/platform/http"
)

type InvitationHandler struct {
	service *InvitationService
}

func NewInvitationHandler(service *InvitationService) *InvitationHandler {
	return &InvitationHandler{service: service}
}

func (h *InvitationHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := http2.ClaimsFromContext(r.Context())
	if !ok || claims == nil || claims.UserID <= 0 {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}
	var input struct {
		ExpiresAt time.Time `json:"expiresAt"`
		MaxUses   int       `json:"maxUses"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	var extra json.RawMessage
	if decoder.Decode(&input) != nil || decoder.Decode(&extra) != io.EOF {
		http2.RespondError(w, &apperrors.ValidationError{Field: "body"})
		return
	}
	invite, err := h.service.Create(r.Context(), claims.UserID, input.ExpiresAt, input.MaxUses)
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	http2.RespondJSON(w, http.StatusCreated, invite)
}

func (h *InvitationHandler) List(w http.ResponseWriter, r *http.Request) {
	invites, err := h.service.List(r.Context())
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, invites)
}

func (h *InvitationHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil || id <= 0 {
		http2.RespondError(w, &apperrors.ValidationError{Field: "id"})
		return
	}
	if err := h.service.Revoke(r.Context(), int(id)); err != nil {
		http2.RespondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *InvitationHandler) Preview(w http.ResponseWriter, r *http.Request) {
	claims, ok := http2.ClaimsFromContext(r.Context())
	if !ok || claims == nil || claims.UserID <= 0 {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}
	var input struct {
		Code string `json:"code"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	var extra json.RawMessage
	if decoder.Decode(&input) != nil || decoder.Decode(&extra) != io.EOF {
		http2.RespondError(w, &apperrors.ValidationError{Field: "body"})
		return
	}
	result, err := h.service.Preview(r.Context(), claims.UserID, input.Code)
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, result)
}

func (h *InvitationHandler) Redeem(w http.ResponseWriter, r *http.Request) {
	claims, ok := http2.ClaimsFromContext(r.Context())
	if !ok || claims == nil || claims.UserID <= 0 {
		http2.RespondError(w, &apperrors.UnauthorizedError{})
		return
	}
	var input struct {
		Code string `json:"code"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	var extra json.RawMessage
	if decoder.Decode(&input) != nil || decoder.Decode(&extra) != io.EOF {
		http2.RespondError(w, &apperrors.ValidationError{Field: "body"})
		return
	}
	result, err := h.service.Redeem(r.Context(), claims.UserID, input.Code)
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, result)
}
