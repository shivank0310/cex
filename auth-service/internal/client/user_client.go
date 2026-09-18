package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/shivank0310/cex.git/auth-service/internal/apperrors"
)

// UserProfile is the user record stored by user-service (no credentials).
type UserProfile struct {
	ID        string
	Email     string
	Username  string
	Status    string
	KYCStatus string
}

type UserClient interface {
	CreateUser(ctx context.Context, email, username string) (UserProfile, error)
	GetByEmail(ctx context.Context, email string) (UserProfile, error)
	GetByID(ctx context.Context, id string) (UserProfile, error)
}

type HTTPUserClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPUserClient(baseURL string) *HTTPUserClient {
	return &HTTPUserClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Status    string `json:"status"`
	KYCStatus string `json:"kyc_status"`
}

func (c *HTTPUserClient) CreateUser(ctx context.Context, email, username string) (UserProfile, error) {
	body, _ := json.Marshal(map[string]string{"email": email, "username": username})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/users", bytes.NewReader(body))
	if err != nil {
		return UserProfile{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doUser(req)
}

func (c *HTTPUserClient) GetByEmail(ctx context.Context, email string) (UserProfile, error) {
	url := fmt.Sprintf("%s/api/v1/users/by-email/%s", c.baseURL, email)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return UserProfile{}, err
	}
	return c.doUser(req)
}

func (c *HTTPUserClient) GetByID(ctx context.Context, id string) (UserProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/users/"+id, nil)
	if err != nil {
		return UserProfile{}, err
	}
	return c.doUser(req)
}

func (c *HTTPUserClient) doUser(req *http.Request) (UserProfile, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return UserProfile{}, apperrors.Wrap(apperrors.CodeInternal, "user-service unavailable", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return UserProfile{}, apperrors.Wrap(apperrors.CodeInternal, "read user-service response", err)
	}
	if resp.StatusCode >= 400 {
		return UserProfile{}, mapUserServiceError(resp.StatusCode, data)
	}

	var payload userResponse
	if err := json.Unmarshal(data, &payload); err != nil {
		return UserProfile{}, apperrors.Wrap(apperrors.CodeInternal, "decode user-service response", err)
	}
	return toProfile(payload), nil
}

func mapUserServiceError(status int, data []byte) error {
	var errBody struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(data, &errBody)
	msg := errBody.Message
	if msg == "" {
		msg = "user-service error"
	}
	switch status {
	case http.StatusConflict:
		return apperrors.New(apperrors.CodeConflict, msg)
	case http.StatusNotFound:
		return apperrors.New(apperrors.CodeNotFound, msg)
	case http.StatusBadRequest:
		return apperrors.New(apperrors.CodeInvalidRequest, msg)
	default:
		return apperrors.New(apperrors.CodeInternal, msg)
	}
}

func toProfile(payload userResponse) UserProfile {
	return UserProfile{
		ID:        payload.ID,
		Email:     payload.Email,
		Username:  payload.Username,
		Status:    payload.Status,
		KYCStatus: payload.KYCStatus,
	}
}

// InMemoryUserClient backs auth-service tests without user-service.
type InMemoryUserClient struct {
	mu      sync.Mutex
	users   map[string]UserProfile
	byEmail map[string]string
	seq     int
}

func NewInMemoryUserClient() *InMemoryUserClient {
	return &InMemoryUserClient{
		users:   make(map[string]UserProfile),
		byEmail: make(map[string]string),
	}
}

func (c *InMemoryUserClient) CreateUser(ctx context.Context, email, username string) (UserProfile, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	email = strings.ToLower(strings.TrimSpace(email))
	if _, ok := c.byEmail[email]; ok {
		return UserProfile{}, apperrors.New(apperrors.CodeConflict, "email already registered")
	}
	c.seq++
	id := fmt.Sprintf("user-%d", c.seq)
	profile := UserProfile{
		ID: id, Email: email, Username: username, Status: "ACTIVE", KYCStatus: "NONE",
	}
	c.users[id] = profile
	c.byEmail[email] = id
	return profile, nil
}

func (c *InMemoryUserClient) GetByEmail(ctx context.Context, email string) (UserProfile, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	id, ok := c.byEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return UserProfile{}, apperrors.New(apperrors.CodeNotFound, "user not found")
	}
	return c.users[id], nil
}

func (c *InMemoryUserClient) GetByID(ctx context.Context, id string) (UserProfile, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	profile, ok := c.users[id]
	if !ok {
		return UserProfile{}, apperrors.New(apperrors.CodeNotFound, "user not found")
	}
	return profile, nil
}
