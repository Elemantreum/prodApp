package section

import (
	"context"
	"net/http"
	"strconv"

	"prodApp/internal/platform/apperrors"
	http2 "prodApp/internal/platform/http"
)

type AccessRepository interface {
	FindAccess(ctx context.Context, userID, sectionID int) (admin, allowed bool, err error)
	WithLockedSection(ctx context.Context, id int, apply func(LockedSectionRepository) error) error
}

type LockedSectionRepository interface {
	SectionWriter
	IsAdmin(ctx context.Context, userID int) (bool, error)
}

type SectionAccess struct {
	repo AccessRepository
}

func NewSectionAccess(repo AccessRepository) *SectionAccess {
	return &SectionAccess{repo: repo}
}

func (a *SectionAccess) Permissions(scoped bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := http2.ClaimsFromContext(r.Context())
			if !ok {
				http2.RespondError(w, &apperrors.UnauthorizedError{})
				return
			}
			id := defaultSectionID
			if input := r.Header.Get("X-Section-ID"); scoped && input != "" {
				var err error
				id, err = strconv.Atoi(input)
				if err != nil || id < 1 {
					http2.RespondError(w, &apperrors.ValidationError{Field: "sectionId"})
					return
				}
			}
			accessSectionID := 0
			if scoped {
				accessSectionID = id
			}
			admin, allowed, err := a.repo.FindAccess(r.Context(), claims.UserID, accessSectionID)
			if err != nil {
				http2.RespondError(w, err)
				return
			}
			if !allowed {
				http2.RespondError(w, &apperrors.ForbiddenError{})
				return
			}
			live := *claims
			live.IsAdmin = admin
			next.ServeHTTP(w, r.WithContext(http2.ContextWithClaims(WithID(r.Context(), id), &live)))
		})
	}
}

func (a *SectionAccess) WithSectionAdmin(ctx context.Context, sectionID int, claims *http2.Claims, apply func(SectionWriter) error) error {
	return a.repo.WithLockedSection(ctx, sectionID, func(repo LockedSectionRepository) error {
		if !claims.IsAdmin {
			admin, err := repo.IsAdmin(ctx, claims.UserID)
			if err != nil {
				return err
			}
			if !admin {
				return &apperrors.ForbiddenError{}
			}
		}
		return apply(repo)
	})
}

func (a *SectionAccess) WithMembershipAccess(ctx context.Context, change MembershipChange, claims *http2.Claims, apply func(SectionWriter) error) error {
	return a.repo.WithLockedSection(ctx, change.SectionID, func(repo LockedSectionRepository) error {
		if !claims.IsAdmin {
			if change.Admin {
				return &apperrors.ForbiddenError{}
			}
			admin, err := repo.IsAdmin(ctx, claims.UserID)
			if err != nil {
				return err
			}
			if !admin {
				return &apperrors.ForbiddenError{}
			}
			protected, err := repo.IsAdmin(ctx, change.UserID)
			if err != nil {
				return err
			}
			if protected {
				return &apperrors.ForbiddenError{}
			}
		}
		return apply(repo)
	})
}
