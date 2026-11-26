package sectioninvite

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"math"
	"strings"
	"time"

	"prodApp/internal/platform/apperrors"
	"prodApp/internal/section"
)

type InvitationRepository interface {
	Create(ctx context.Context, sectionID, createdBy int, tokenHash string, expiresAt time.Time, maxUses int) (Invitation, error)
	List(ctx context.Context, sectionID int) ([]Invitation, error)
	Revoke(ctx context.Context, sectionID, inviteID int) error
	Lookup(ctx context.Context, tokenHash string, userID int) (InviteState, error)
	WithLockedInvite(ctx context.Context, tokenHash string, userID int, apply func(InviteState, bool, InviteWriter) error) error
}

type InviteWriter interface {
	InsertMember(ctx context.Context) (bool, error)
	Consume(ctx context.Context) error
}

type InviteState struct {
	Invitation Invitation
	Section    section.Section
	Member     bool
}

type InvitationService struct {
	repo InvitationRepository
}

func NewInvitationService(repo InvitationRepository) *InvitationService {
	return &InvitationService{repo: repo}
}

func (s *InvitationService) Create(ctx context.Context, createdBy int, expiresAt time.Time, maxUses int) (Invitation, error) {
	expiresAt = expiresAt.UTC()
	if !expiresAt.After(time.Now()) || expiresAt.Year() > 9999 {
		return Invitation{}, &apperrors.ValidationError{Field: "expiresAt"}
	}
	if maxUses < 1 || maxUses > math.MaxInt32 {
		return Invitation{}, &apperrors.ValidationError{Field: "maxUses"}
	}
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return Invitation{}, err
	}
	code := base64.RawURLEncoding.EncodeToString(token[:])
	invite, err := s.repo.Create(ctx, section.ID(ctx), createdBy, hashCode(code), expiresAt, maxUses)
	if err != nil {
		return Invitation{}, err
	}
	invite.Code = code
	invite.normalizeTimes()
	return invite, nil
}

func (s *InvitationService) List(ctx context.Context) ([]Invitation, error) {
	invites, err := s.repo.List(ctx, section.ID(ctx))
	if err != nil {
		return nil, err
	}
	if len(invites) == 0 {
		return []Invitation{}, nil
	}
	for i := range invites {
		invites[i].normalizeTimes()
	}
	return invites, nil
}

func (s *InvitationService) Revoke(ctx context.Context, inviteID int) error {
	return s.repo.Revoke(ctx, section.ID(ctx), inviteID)
}

func (s *InvitationService) Preview(ctx context.Context, userID int, code string) (section.Section, error) {
	if strings.TrimSpace(code) == "" {
		return section.Section{}, &apperrors.ValidationError{Field: "code"}
	}
	state, err := s.repo.Lookup(ctx, hashCode(code), userID)
	if err != nil {
		return section.Section{}, err
	}
	if err := state.checkAvailability(time.Now()); err != nil {
		return section.Section{}, err
	}
	return state.Section, nil
}

func (s *InvitationService) Redeem(ctx context.Context, userID int, code string) (section.Section, error) {
	if strings.TrimSpace(code) == "" {
		return section.Section{}, &apperrors.ValidationError{Field: "code"}
	}
	var joined section.Section
	err := s.repo.WithLockedInvite(ctx, hashCode(code), userID, func(state InviteState, registered bool, writer InviteWriter) error {
		if !registered {
			return &apperrors.ForbiddenError{Reason: "registration required"}
		}
		if err := state.checkAvailability(time.Now()); err != nil {
			return err
		}
		joined = state.Section
		if state.Member {
			return nil
		}
		inserted, err := writer.InsertMember(ctx)
		if err != nil || !inserted {
			return err
		}
		return writer.Consume(ctx)
	})
	if err != nil {
		return section.Section{}, err
	}
	return joined, nil
}

func (s InviteState) checkAvailability(now time.Time) error {
	if s.Member {
		return nil
	}
	const resource = "invitation"
	switch {
	case s.Invitation.RevokedAt != nil:
		return &apperrors.ConflictError{Resource: resource, Reason: "revoked"}
	case !now.Before(s.Invitation.ExpiresAt):
		return &apperrors.ConflictError{Resource: resource, Reason: "expired"}
	case s.Invitation.Uses >= s.Invitation.MaxUses:
		return &apperrors.ConflictError{Resource: resource, Reason: "exhausted"}
	default:
		return nil
	}
}

func hashCode(code string) string {
	digest := sha256.Sum256([]byte(code))
	return hex.EncodeToString(digest[:])
}
