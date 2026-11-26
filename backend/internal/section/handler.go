package section

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"prodApp/internal/platform/apperrors"
	http2 "prodApp/internal/platform/http"
)

type SectionHandler struct {
	service *SectionService
	access  *SectionAccess
}

func NewSectionHandler(service *SectionService, access *SectionAccess) *SectionHandler {
	return &SectionHandler{service: service, access: access}
}

func (h *SectionHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := http2.ClaimsFromContext(r.Context())
	sections, err := h.service.List(r.Context(), claims.UserID)
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, Directory{Sections: sections, SystemAdmin: claims.IsAdmin})
}

func (h *SectionHandler) Save(w http.ResponseWriter, r *http.Request) {
	var input Section
	r.Body = http.MaxBytesReader(w, r.Body, 400*1024)
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "name"})
		return
	}
	claims, _ := http2.ClaimsFromContext(r.Context())
	if input.ID == 0 {
		if !claims.IsAdmin {
			http2.RespondError(w, &apperrors.ForbiddenError{})
			return
		}
		saved, err := h.service.Create(r.Context(), input)
		if err != nil {
			http2.RespondError(w, err)
			return
		}
		http2.RespondJSON(w, http.StatusOK, saved)
		return
	}
	err := h.access.WithSectionAdmin(r.Context(), input.ID, claims, func(writer SectionWriter) error {
		return h.service.Update(r.Context(), writer, input)
	})
	var missing *apperrors.NotFoundError
	if errors.As(err, &missing) {
		http2.RespondError(w, &apperrors.ForbiddenError{})
		return
	}
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, input)
}

func (h *SectionHandler) ChangeMembership(w http.ResponseWriter, r *http.Request) {
	var input MembershipChange
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.SectionID < 1 || input.UserID < 1 {
		http2.RespondError(w, &apperrors.ValidationError{Field: "member"})
		return
	}
	claims, _ := http2.ClaimsFromContext(r.Context())
	err := h.access.WithMembershipAccess(r.Context(), input, claims, func(writer SectionWriter) error {
		return h.service.ChangeMembership(r.Context(), writer, input)
	})
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, true)
}

func (h *SectionHandler) Roster(w http.ResponseWriter, r *http.Request) {
	members, err := h.service.Roster(r.Context())
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, members)
}

func (h *SectionHandler) Candidates(w http.ResponseWriter, r *http.Request) {
	members, err := h.service.Candidates(r.Context())
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, members)
}

func (h *SectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http2.RespondError(w, &apperrors.ValidationError{Field: "id"})
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusOK, true)
}

func (h *SectionHandler) AddVenue(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http2.RespondError(w, &apperrors.ValidationError{Field: "name"})
		return
	}
	name, err := h.service.AddVenue(r.Context(), input.Name)
	if err != nil {
		http2.RespondError(w, err)
		return
	}
	http2.RespondJSON(w, http.StatusCreated, name)
}

func (h *SectionHandler) DeleteVenue(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteVenue(r.Context(), r.URL.Query().Get("name")); err != nil {
		http2.RespondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
