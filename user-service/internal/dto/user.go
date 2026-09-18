package dto

import (
	"time"

	"github.com/shivank0310/cex.git/user-service/internal/model"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type UpdateUserRequest struct {
	Username  *string `json:"username"`
	Status    *string `json:"status"`
	KYCStatus *string `json:"kyc_status"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	KYCStatus string    `json:"kyc_status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToUser(u model.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Status:    string(u.Status),
		KYCStatus: string(u.KYCStatus),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
