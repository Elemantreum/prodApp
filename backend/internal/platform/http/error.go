package http

import (
	"errors"
	"log"
	"net/http"
	"prodApp/internal/platform/apperrors"
)

var fb *apperrors.ForbiddenError
var nf apperrors.NotFoundError
var vd *apperrors.ValidationError
var ua *apperrors.UnauthorizedError
var ce *apperrors.ConflictError

func RespondError(w http.ResponseWriter, err error) {
	switch {
	case errors.As(err, &nf):
		RespondJSON(w, http.StatusNotFound, map[string]string{"error": nf.Error()})
	case errors.As(err, &vd):
		RespondJSON(w, http.StatusBadRequest, map[string]string{"error": vd.Error()})
	case errors.As(err, &fb):
		RespondJSON(w, http.StatusForbidden, map[string]string{"error": fb.Error()})
	case errors.As(err, &ua):
		RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": ua.Error()})
	case errors.As(err, &ce):
		RespondJSON(w, http.StatusConflict, map[string]string{"error": ce.Error()})

	default:
		log.Printf("unexpected error: %v", err)
		RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}
