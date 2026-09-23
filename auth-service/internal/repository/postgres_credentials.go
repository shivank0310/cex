package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/shivank0310/cex.git/auth-service/internal/model"
)

type PostgresCredentialStore struct {
	db *sql.DB
}

func NewPostgresCredentialStore(db *sql.DB) *PostgresCredentialStore {
	return &PostgresCredentialStore{db: db}
}

const insertCredentialSQL = `
INSERT INTO auth.credentials (user_id, email, password_hash, role, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING user_id, email, password_hash, role, status, created_at, updated_at`

const selectCredentialByIDSQL = `
SELECT user_id, email, password_hash, role, status, created_at, updated_at
FROM auth.credentials WHERE user_id = $1`

const selectCredentialByEmailSQL = `
SELECT user_id, email, password_hash, role, status, created_at, updated_at
FROM auth.credentials WHERE email = $1`

func (r *PostgresCredentialStore) CreateWithID(id, email, passwordHash string, role model.Role) (model.User, error) {
	email = normalizeEmail(email)
	if id == "" || email == "" {
		return model.User{}, errors.New("id and email required")
	}

	now := time.Now().UTC()
	row := r.db.QueryRow(insertCredentialSQL,
		id, email, passwordHash, string(role), string(model.UserActive), now, now,
	)
	user, err := scanCredential(row)
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, errors.New("email already registered")
		}
		return model.User{}, err
	}
	return user, nil
}

func (r *PostgresCredentialStore) GetByID(id string) (*model.User, bool) {
	row := r.db.QueryRow(selectCredentialByIDSQL, id)
	user, err := scanCredential(row)
	if err != nil {
		return nil, false
	}
	return &user, true
}

func (r *PostgresCredentialStore) GetByEmail(email string) (*model.User, bool) {
	row := r.db.QueryRow(selectCredentialByEmailSQL, normalizeEmail(email))
	user, err := scanCredential(row)
	if err != nil {
		return nil, false
	}
	return &user, true
}

func scanCredential(row interface {
	Scan(dest ...any) error
}) (model.User, error) {
	var user model.User
	var role, status string
	err := row.Scan(
		&user.ID, &user.Email, &user.PasswordHash, &role, &status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return model.User{}, err
	}
	user.Role = model.Role(role)
	user.Status = model.UserStatus(status)
	return user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}
