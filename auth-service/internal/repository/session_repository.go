package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/shivank0310/cex.git/auth-service/internal/model"
)

// MemorySessionStore is for unit tests and local dev without Redis.
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*model.Session // refresh token → session
	byUser   map[string][]string
	seq      int
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]*model.Session),
		byUser:   make(map[string][]string),
	}
}

func (r *MemorySessionStore) Create(userID, refreshToken string, expiresAt time.Time) model.Session {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	session := model.Session{
		ID:           formatSessionID(r.seq),
		UserID:       userID,
		RefreshToken: refreshToken,
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    expiresAt,
	}
	r.sessions[refreshToken] = &session
	r.byUser[userID] = append(r.byUser[userID], refreshToken)
	return session
}

func (r *MemorySessionStore) GetByRefreshToken(token string) (*model.Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[token]
	if !ok {
		return nil, false
	}
	copy := *session
	return &copy, true
}

func (r *MemorySessionStore) Revoke(refreshToken string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.sessions[refreshToken]
	if !ok || session.RevokedAt != nil {
		return false
	}
	now := time.Now().UTC()
	session.RevokedAt = &now
	return true
}

func (r *MemorySessionStore) RevokeAllForUser(userID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	tokens := r.byUser[userID]
	now := time.Now().UTC()
	count := 0
	for _, token := range tokens {
		if session, ok := r.sessions[token]; ok && session.RevokedAt == nil {
			session.RevokedAt = &now
			count++
		}
	}
	return count
}

func formatSessionID(seq int) string {
	return fmt.Sprintf("sess-%d", seq)
}
