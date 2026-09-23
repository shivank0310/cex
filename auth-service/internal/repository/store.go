package repository

import (
	"time"

	"github.com/shivank0310/cex.git/auth-service/internal/model"
)

// CredentialStore persists login identities (password hashes).
type CredentialStore interface {
	CreateWithID(id, email, passwordHash string, role model.Role) (model.User, error)
	GetByID(id string) (*model.User, bool)
	GetByEmail(email string) (*model.User, bool)
}

// SessionStore persists refresh-token sessions.
type SessionStore interface {
	Create(userID, refreshToken string, expiresAt time.Time) model.Session
	GetByRefreshToken(token string) (*model.Session, bool)
	Revoke(refreshToken string) bool
}
