package section

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"strings"
	"unicode/utf8"

	"prodApp/internal/platform/apperrors"
)

type SectionRepository interface {
	List(ctx context.Context, userID int) ([]SectionWithVenues, error)
	Create(ctx context.Context, input Section) (Section, error)
	Roster(ctx context.Context, id int) ([]Member, error)
	Candidates(ctx context.Context, id int) ([]Member, error)
	DeleteEmpty(ctx context.Context, id int) (bool, error)
	AddVenue(ctx context.Context, sectionID int, name string) (bool, error)
	DeleteVenue(ctx context.Context, sectionID int, name string) error
}

type SectionWriter interface {
	UpdateDetails(ctx context.Context, name, logo string) error
	PutMember(ctx context.Context, userID int, admin bool) error
	RemoveMember(ctx context.Context, userID int) error
	RemoveWaitingEntries(ctx context.Context, userID int) error
	CancelUpcomingBookings(ctx context.Context, userID int) error
}

type SectionService struct {
	repo SectionRepository
}

func NewSectionService(repo SectionRepository) *SectionService {
	return &SectionService{repo: repo}
}

func (s *SectionService) List(ctx context.Context, userID int) ([]SectionWithVenues, error) {
	return s.repo.List(ctx, userID)
}

func (s *SectionService) Create(ctx context.Context, input Section) (Section, error) {
	if err := validateSection(input); err != nil {
		return Section{}, err
	}
	return s.repo.Create(ctx, input)
}

func (s *SectionService) Update(ctx context.Context, writer SectionWriter, input Section) error {
	if err := validateSection(input); err != nil {
		return err
	}
	return writer.UpdateDetails(ctx, input.Name, input.Logo)
}

func (s *SectionService) ChangeMembership(ctx context.Context, writer SectionWriter, change MembershipChange) error {
	if !change.Remove {
		return writer.PutMember(ctx, change.UserID, change.Admin)
	}
	if err := writer.RemoveWaitingEntries(ctx, change.UserID); err != nil {
		return err
	}
	if err := writer.CancelUpcomingBookings(ctx, change.UserID); err != nil {
		return err
	}
	return writer.RemoveMember(ctx, change.UserID)
}

func (s *SectionService) Roster(ctx context.Context) ([]Member, error) {
	return s.repo.Roster(ctx, ID(ctx))
}

func (s *SectionService) Candidates(ctx context.Context) ([]Member, error) {
	return s.repo.Candidates(ctx, ID(ctx))
}

func (s *SectionService) Delete(ctx context.Context, id int) error {
	if id == defaultSectionID {
		return &apperrors.ConflictError{Resource: "секция", Reason: "исходную секцию нельзя удалить"}
	}
	deleted, err := s.repo.DeleteEmpty(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return &apperrors.ConflictError{Resource: "секция", Reason: "остались тренировки или серии"}
	}
	return nil
}

func (s *SectionService) AddVenue(ctx context.Context, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 120 || strings.ContainsRune(name, 0) {
		return "", &apperrors.ValidationError{Field: "name"}
	}
	added, err := s.repo.AddVenue(ctx, ID(ctx), name)
	if err != nil {
		return "", err
	}
	if !added {
		return "", &apperrors.ConflictError{Resource: "место проведения", Reason: "такое название уже есть в списке"}
	}
	return name, nil
}

func (s *SectionService) DeleteVenue(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" || strings.ContainsRune(name, 0) {
		return &apperrors.ValidationError{Field: "name"}
	}
	return s.repo.DeleteVenue(ctx, ID(ctx), name)
}

func validateSection(input Section) error {
	if strings.TrimSpace(input.Name) == "" || utf8.RuneCountInString(input.Name) > 120 {
		return &apperrors.ValidationError{Field: "name"}
	}
	if !validLogo(input.Logo) {
		return &apperrors.ValidationError{Field: "logo"}
	}
	return nil
}

func validLogo(value string) bool {
	if value == "" {
		return true
	}
	if strings.HasPrefix(value, "data:") {
		header, payload, ok := strings.Cut(value, ",")
		if !ok || len(payload) > base64.StdEncoding.EncodedLen(256*1024) {
			return false
		}
		if header != "data:image/png;base64" && header != "data:image/jpeg;base64" {
			return false
		}
		raw, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return false
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
		return err == nil && header == "data:image/"+format+";base64" && cfg.Width <= 4096 && cfg.Height <= 4096
	}
	u, err := url.Parse(value)
	return err == nil && u.Scheme == "https" && u.Host != "" && len(value) <= 2048
}
