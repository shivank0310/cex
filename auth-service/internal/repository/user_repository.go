package repository

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/shivank0310/cex.git/auth-service/internal/model"
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*model.User
	byEmail map[string]string
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:   make(map[string]*model.User),
		byEmail: make(map[string]string),
	}
}

func (r *UserRepository) CreateWithID(id, email, passwordHash string, role model.Role) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	email = strings.ToLower(strings.TrimSpace(email))
	if id == "" || email == "" {
		return model.User{}, errors.New("id and email required")
	}
	if _, exists := r.byEmail[email]; exists {
		return model.User{}, errors.New("email already registered")
	}

	now := time.Now().UTC()
	user := model.User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		Status:       model.UserActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	r.users[user.ID] = &user
	r.byEmail[email] = user.ID
	return user, nil
}

func (r *UserRepository) GetByID(id string) (*model.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return nil, false
	}
	copy := *user
	return &copy, true
}

func (r *UserRepository) GetByEmail(email string) (*model.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return nil, false
	}
	user := r.users[id]
	copy := *user
	return &copy, true
}
