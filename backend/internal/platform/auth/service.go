package auth

import (
	"context"
	"errors"
	"time"

	"prodApp/internal/platform/apperrors"
	"prodApp/internal/user"
)

type AccountService interface {
	EnsureProfile(ctx context.Context, provider, externalID, firstName, lastName, username string) (*user.User, error)
	ResolveAndLink(ctx context.Context, code, provider, externalID, username string) (*user.User, error)
	GetUser(ctx context.Context, userID int) (*user.User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, userID int, ttl time.Duration) (string, error)
	FindActiveUserID(ctx context.Context, refreshToken string) (int, error)
	Revoke(ctx context.Context, refreshToken string) error
}

type AuthService struct {
	issuer   *TokenIssuer
	sessions SessionRepository
	accounts AccountService
}

func NewAuthService(issuer *TokenIssuer, sessions SessionRepository, accounts AccountService) *AuthService {
	return &AuthService{issuer: issuer, sessions: sessions, accounts: accounts}
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *AuthService) Login(ctx context.Context, principal *Principal) (tokenResponse, error) {
	account, err := s.accounts.EnsureProfile(ctx, principal.Provider, principal.ExternalId, principal.FirstName, principal.LastName, principal.Username)
	if err != nil {
		return tokenResponse{}, err
	}
	return s.issueTokenPair(ctx, account)
}

func (s *AuthService) Link(ctx context.Context, principal *Principal, code string) (tokenResponse, error) {
	if code == "" {
		return tokenResponse{}, &apperrors.ValidationError{Field: "code"}
	}
	account, err := s.accounts.ResolveAndLink(ctx, code, principal.Provider, principal.ExternalId, principal.Username)
	if err != nil {
		return tokenResponse{}, err
	}
	return s.issueTokenPair(ctx, account)
}

func (s *AuthService) Refresh(ctx context.Context, token string) (tokenResponse, error) {
	userID, err := s.sessions.FindActiveUserID(ctx, token)
	if errors.Is(err, ErrRefreshTokenInvalid) {
		return tokenResponse{}, &apperrors.UnauthorizedError{}
	}
	if err != nil {
		return tokenResponse{}, err
	}
	if err := s.sessions.Revoke(ctx, token); err != nil {
		return tokenResponse{}, err
	}
	account, err := s.accounts.GetUser(ctx, userID)
	if err != nil {
		return tokenResponse{}, err
	}
	return s.issueTokenPair(ctx, account)
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessions.Revoke(ctx, token)
}

func (s *AuthService) issueTokenPair(ctx context.Context, account *user.User) (tokenResponse, error) {
	access, err := s.issuer.IssueAccessToken(account.Id, account.IsAdmin)
	if err != nil {
		return tokenResponse{}, err
	}
	refresh, err := s.sessions.Create(ctx, account.Id, s.issuer.RefreshTokenTTL())
	if err != nil {
		return tokenResponse{}, err
	}
	return tokenResponse{AccessToken: access, RefreshToken: refresh}, nil
}
