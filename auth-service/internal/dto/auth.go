package dto

import (
	"time"

	"github.com/shivank0310/cex.git/auth-service/internal/model"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type AuthResponse struct {
	User  UserResponse  `json:"user"`
	Token TokenResponse `json:"token"`
}

type UserResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	Status       string `json:"status"`
	KYCStatus    string `json:"kyc_status"`
	TwoFAEnabled bool   `json:"two_fa_enabled"`
}

func ToUser(u model.User) UserResponse {
	return UserResponse{
		ID:           u.ID,
		Email:        u.Email,
		Username:     u.Username,
		Role:         string(u.Role),
		Status:       string(u.Status),
		KYCStatus:    u.KYCStatus,
		TwoFAEnabled: u.TwoFAEnabled,
	}
}

func ToToken(pair model.TokenPair) TokenResponse {
	return TokenResponse{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        pair.TokenType,
		AccessExpiresAt:  pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
	}
}

func ToAuthResponse(user model.User, pair model.TokenPair) AuthResponse {
	return AuthResponse{
		User:  ToUser(user),
		Token: ToToken(pair),
	}
}
