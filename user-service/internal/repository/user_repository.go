package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shivank0310/cex.git/user-service/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

const insertUserSQL = `
INSERT INTO users.accounts (id, email, username, status, kyc_status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, email, COALESCE(username, ''), status, kyc_status, created_at, updated_at`

const selectByIDSQL = `
SELECT id, email, COALESCE(username, ''), status, kyc_status, created_at, updated_at
FROM users.accounts WHERE id = $1`

const selectByEmailSQL = `
SELECT id, email, COALESCE(username, ''), status, kyc_status, created_at, updated_at
FROM users.accounts WHERE email = $1`

const updateUserSQL = `
UPDATE users.accounts
SET username = COALESCE($2, username),
    status = COALESCE($3, status),
    kyc_status = COALESCE($4, kyc_status),
    updated_at = $5
WHERE id = $1
RETURNING id, email, COALESCE(username, ''), status, kyc_status, created_at, updated_at`

func (r *UserRepository) Create(ctx context.Context, email, username string) (model.User, error) {
	email = normalizeEmail(email)
	if email == "" {
		return model.User{}, errors.New("email required")
	}
	username = strings.TrimSpace(username)
	now := time.Now().UTC()
	id := newUserID()

	row := r.db.QueryRowContext(ctx, insertUserSQL,
		id, email, nullableString(username), string(model.StatusActive), string(model.KYCNone), now, now,
	)
	user, err := scanUser(row)
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, errors.New("email or username already exists")
		}
		return model.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (model.User, error) {
	row := r.db.QueryRowContext(ctx, selectByIDSQL, id)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, errors.New("user not found")
		}
		return model.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	row := r.db.QueryRowContext(ctx, selectByEmailSQL, normalizeEmail(email))
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, errors.New("user not found")
		}
		return model.User{}, err
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, id string, username *string, status *model.Status, kycStatus *model.KYCStatus) (model.User, error) {
	var statusVal, kycVal interface{}
	if status != nil {
		statusVal = string(*status)
	}
	if kycStatus != nil {
		kycVal = string(*kycStatus)
	}
	row := r.db.QueryRowContext(ctx, updateUserSQL, id, username, statusVal, kycVal, time.Now().UTC())
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, errors.New("user not found")
		}
		if isUniqueViolation(err) {
			return model.User{}, errors.New("username already exists")
		}
		return model.User{}, err
	}
	return user, nil
}

func scanUser(row interface {
	Scan(dest ...any) error
}) (model.User, error) {
	var user model.User
	err := row.Scan(
		&user.ID, &user.Email, &user.Username, &user.Status, &user.KYCStatus, &user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func newUserID() string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("user-%s", hex.EncodeToString(buf))
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}
