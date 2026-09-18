package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/shivank0310/cex.git/auth-service/internal/apperrors"
	"github.com/shivank0310/cex.git/auth-service/internal/client"
	"github.com/shivank0310/cex.git/auth-service/internal/config"
	"github.com/shivank0310/cex.git/auth-service/internal/model"
	"github.com/shivank0310/cex.git/auth-service/internal/repository"
	"github.com/shivank0310/cex.git/pkg/jwt"
)

type AuthService struct {
	cfg      config.Config
	users    *repository.UserRepository
	sessions *repository.SessionRepository
	profiles client.UserClient
	tokens   *jwt.Manager
}

func NewAuthService(cfg config.Config, users *repository.UserRepository, sessions *repository.SessionRepository, profiles client.UserClient) *AuthService {
	return &AuthService{
		cfg:      cfg,
		users:    users,
		sessions: sessions,
		profiles: profiles,
		tokens:   jwt.NewManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.AccessTokenTTL),
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, username string) (model.User, model.TokenPair, error) {
	if err := validateCredentials(email, password); err != nil {
		return model.User{}, model.TokenPair{}, err
	}

	profile, err := s.profiles.CreateUser(ctx, email, username)
	if err != nil {
		return model.User{}, model.TokenPair{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.BcryptCost)
	if err != nil {
		return model.User{}, model.TokenPair{}, apperrors.Wrap(apperrors.CodeInternal, "password hash failed", err)
	}

	user, err := s.users.CreateWithID(profile.ID, email, string(hash), model.RoleTrader)
	if err != nil {
		if strings.Contains(err.Error(), "already registered") {
			return model.User{}, model.TokenPair{}, apperrors.New(apperrors.CodeConflict, "email already registered")
		}
		return model.User{}, model.TokenPair{}, apperrors.Wrap(apperrors.CodeInternal, "create credentials failed", err)
	}
	user = s.enrichUser(user, profile)

	pair, err := s.issueTokens(user)
	if err != nil {
		return model.User{}, model.TokenPair{}, err
	}
	return user, pair, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (model.User, model.TokenPair, error) {
	user, ok := s.users.GetByEmail(email)
	if !ok {
		return model.User{}, model.TokenPair{}, apperrors.New(apperrors.CodeUnauthorized, "invalid email or password")
	}
	if user.Status != model.UserActive {
		return model.User{}, model.TokenPair{}, apperrors.New(apperrors.CodeUnauthorized, "account suspended")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return model.User{}, model.TokenPair{}, apperrors.New(apperrors.CodeUnauthorized, "invalid email or password")
	}

	authUser := *user
	profile, err := s.profiles.GetByEmail(ctx, email)
	if err == nil {
		if profile.Status == "SUSPENDED" {
			return model.User{}, model.TokenPair{}, apperrors.New(apperrors.CodeUnauthorized, "account suspended")
		}
		authUser = s.enrichUser(authUser, profile)
	}

	pair, err := s.issueTokens(authUser)
	if err != nil {
		return model.User{}, model.TokenPair{}, err
	}
	return authUser, pair, nil
}

func (s *AuthService) Refresh(refreshToken string) (model.User, model.TokenPair, error) {
	session, ok := s.sessions.GetByRefreshToken(refreshToken)
	if !ok || !session.Active(time.Now().UTC()) {
		return model.User{}, model.TokenPair{}, apperrors.New(apperrors.CodeUnauthorized, "invalid refresh token")
	}

	user, ok := s.users.GetByID(session.UserID)
	if !ok || user.Status != model.UserActive {
		return model.User{}, model.TokenPair{}, apperrors.New(apperrors.CodeUnauthorized, "invalid refresh token")
	}

	s.sessions.Revoke(refreshToken)
	pair, err := s.createTokenPair(*user)
	if err != nil {
		return model.User{}, model.TokenPair{}, err
	}
	return *user, pair, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	if refreshToken == "" {
		return apperrors.New(apperrors.CodeInvalidRequest, "refresh_token required")
	}
	if !s.sessions.Revoke(refreshToken) {
		return apperrors.New(apperrors.CodeUnauthorized, "invalid refresh token")
	}
	return nil
}

func (s *AuthService) Me(ctx context.Context, accessToken string) (model.User, error) {
	claims, err := s.tokens.ValidateAccess(accessToken)
	if err != nil {
		return model.User{}, apperrors.New(apperrors.CodeUnauthorized, "invalid access token")
	}
	user, ok := s.users.GetByID(claims.Subject)
	if !ok || user.Status != model.UserActive {
		return model.User{}, apperrors.New(apperrors.CodeUnauthorized, "invalid access token")
	}

	profile, err := s.profiles.GetByID(ctx, user.ID)
	if err == nil {
		return s.enrichUser(*user, profile), nil
	}
	return *user, nil
}

func (s *AuthService) JWTManager() *jwt.Manager {
	return s.tokens
}

func (s *AuthService) enrichUser(user model.User, profile client.UserProfile) model.User {
	user.Username = profile.Username
	user.KYCStatus = profile.KYCStatus
	if profile.Status == "SUSPENDED" {
		user.Status = model.UserSuspended
	}
	return user
}

func (s *AuthService) issueTokens(user model.User) (model.TokenPair, error) {
	return s.createTokenPair(user)
}

func (s *AuthService) createTokenPair(user model.User) (model.TokenPair, error) {
	refreshToken, err := newRefreshToken()
	if err != nil {
		return model.TokenPair{}, apperrors.Wrap(apperrors.CodeInternal, "refresh token failed", err)
	}
	refreshExpires := time.Now().UTC().Add(s.cfg.RefreshTokenTTL)
	session := s.sessions.Create(user.ID, refreshToken, refreshExpires)

	accessToken, accessExpires, err := s.tokens.SignAccess(user.ID, string(user.Role), session.ID)
	if err != nil {
		return model.TokenPair{}, apperrors.Wrap(apperrors.CodeInternal, "access token failed", err)
	}

	return model.TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpires,
		RefreshExpiresAt: refreshExpires,
		TokenType:        "Bearer",
	}, nil
}

func validateCredentials(email, password string) error {
	email = strings.TrimSpace(email)
	if email == "" || !strings.Contains(email, "@") {
		return apperrors.New(apperrors.CodeInvalidRequest, "valid email required")
	}
	if len(password) < 8 {
		return apperrors.New(apperrors.CodeInvalidRequest, "password must be at least 8 characters")
	}
	return nil
}

func newRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
