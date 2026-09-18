package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shivank0310/cex.git/user-service/internal/model"
)

// MemoryRepository is used for unit tests without PostgreSQL.
type MemoryRepository struct {
	mu        sync.RWMutex
	users     map[string]model.User
	byEmail   map[string]string
	byUsername map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:      make(map[string]model.User),
		byEmail:    make(map[string]string),
		byUsername: make(map[string]string),
	}
}

func (r *MemoryRepository) Create(ctx context.Context, email, username string) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	email = strings.ToLower(strings.TrimSpace(email))
	username = strings.TrimSpace(username)
	if email == "" {
		return model.User{}, errors.New("email required")
	}
	if _, ok := r.byEmail[email]; ok {
		return model.User{}, errors.New("email or username already exists")
	}
	if username != "" {
		if _, ok := r.byUsername[strings.ToLower(username)]; ok {
			return model.User{}, errors.New("email or username already exists")
		}
	}

	now := time.Now().UTC()
	user := model.User{
		ID:        newMemoryUserID(),
		Email:     email,
		Username:  username,
		Status:    model.StatusActive,
		KYCStatus: model.KYCNone,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.users[user.ID] = user
	r.byEmail[email] = user.ID
	if username != "" {
		r.byUsername[strings.ToLower(username)] = user.ID
	}
	return user, nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return model.User{}, errors.New("user not found")
	}
	return user, nil
}

func (r *MemoryRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return model.User{}, errors.New("user not found")
	}
	return r.users[id], nil
}

func (r *MemoryRepository) Update(ctx context.Context, id string, username *string, status *model.Status, kycStatus *model.KYCStatus) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[id]
	if !ok {
		return model.User{}, errors.New("user not found")
	}
	if username != nil {
		name := strings.TrimSpace(*username)
		if name != "" {
			key := strings.ToLower(name)
			if existing, taken := r.byUsername[key]; taken && existing != id {
				return model.User{}, errors.New("username already exists")
			}
			if user.Username != "" {
				delete(r.byUsername, strings.ToLower(user.Username))
			}
			user.Username = name
			r.byUsername[key] = id
		}
	}
	if status != nil {
		user.Status = *status
	}
	if kycStatus != nil {
		user.KYCStatus = *kycStatus
	}
	user.UpdatedAt = time.Now().UTC()
	r.users[id] = user
	return user, nil
}

func newMemoryUserID() string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("user-%s", hex.EncodeToString(buf))
}
