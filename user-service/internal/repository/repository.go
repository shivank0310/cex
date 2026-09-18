package repository

import (
	"context"

	"github.com/shivank0310/cex.git/user-service/internal/model"
)

// UserStore persists user profile records.
type UserStore interface {
	Create(ctx context.Context, email, username string) (model.User, error)
	GetByID(ctx context.Context, id string) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
	Update(ctx context.Context, id string, username *string, status *model.Status, kycStatus *model.KYCStatus) (model.User, error)
}
