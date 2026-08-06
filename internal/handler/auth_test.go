package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockAuthService struct {
	loginFunc    func(ctx context.Context, login, password string) (string, error)
	registerFunc func(ctx context.Context, login, password string) error
}

func (m *MockAuthService) Login(ctx context.Context, login, password string) (string, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, login, password)
	}
	return "", errors.New("not implemented in mock")
}

func (m *MockAuthService) Register(ctx context.Context, login, password string) error {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, login, password)
	}
	return errors.New("not implemented in mock")
}

func TestLoginSuccess(t *testing.T) {
	mockService := &MockAuthService{
		loginFunc: func(ctx context.Context, login, password string) (string, error) {
			if login == "testuser" && password == "testpass" {
				return "valid_token_xyz", nil
			}
			return "", errors.New("invalid credentials")
		},
	}

	handler := NewAuthHandler(mockService)

	reqBody := RegisterRequest{
		Login:    "testuser",
		Password: "testpass",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp LoginResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	if err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if resp.Token != "valid_token_xyz" {
		t.Errorf("expected token 'valid_token_xyz', got %q", resp.Token)
	}
}

func TestLoginInvalidJSON(t *testing.T) {
	mockService := &MockAuthService{}
	handler := NewAuthHandler(mockService)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if !contains(w.Body.String(), "failed to validate request data") {
		t.Errorf("expected error message about validation, got %q", w.Body.String())
	}
}

func TestLoginServiceError(t *testing.T) {
	mockService := &MockAuthService{
		loginFunc: func(ctx context.Context, login, password string) (string, error) {
			return "", errors.New("invalid credentials")
		},
	}

	handler := NewAuthHandler(mockService)

	reqBody := RegisterRequest{
		Login:    "wronguser",
		Password: "wrongpass",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if !contains(w.Body.String(), "failed to login") {
		t.Errorf("expected error message, got %q", w.Body.String())
	}
}

func TestLoginEmptyCredentials(t *testing.T) {
	mockService := &MockAuthService{
		loginFunc: func(ctx context.Context, login, password string) (string, error) {
			if login == "" || password == "" {
				return "", errors.New("empty credentials")
			}
			return "token", nil
		},
	}

	handler := NewAuthHandler(mockService)

	tests := []struct {
		name     string
		login    string
		password string
	}{
		{"empty login", "", "password"},
		{"empty password", "login", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := RegisterRequest{
				Login:    tt.login,
				Password: tt.password,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestLoginResponseContentType(t *testing.T) {
	mockService := &MockAuthService{
		loginFunc: func(ctx context.Context, login, password string) (string, error) {
			return "token", nil
		},
	}

	handler := NewAuthHandler(mockService)

	reqBody := RegisterRequest{
		Login:    "user",
		Password: "pass",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}
}

func TestRegisterSuccess(t *testing.T) {
	registerCalled := false
	mockService := &MockAuthService{
		registerFunc: func(ctx context.Context, login, password string) error {
			registerCalled = true
			if login == "newuser" && password == "newpass" {
				return nil
			}
			return errors.New("registration failed")
		},
	}

	handler := NewAuthHandler(mockService)

	reqBody := RegisterRequest{
		Login:    "newuser",
		Password: "newpass",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if !registerCalled {
		t.Error("expected register service to be called")
	}

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestRegisterInvalidJSON(t *testing.T) {
	mockService := &MockAuthService{}
	handler := NewAuthHandler(mockService)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader([]byte("{invalid}")))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if !contains(w.Body.String(), "failed to validate request data") {
		t.Errorf("expected validation error, got %q", w.Body.String())
	}
}

func TestRegisterServiceError(t *testing.T) {
	mockService := &MockAuthService{
		registerFunc: func(ctx context.Context, login, password string) error {
			return errors.New("login already exists")
		},
	}

	handler := NewAuthHandler(mockService)

	reqBody := RegisterRequest{
		Login:    "existinguser",
		Password: "pass",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	if !contains(w.Body.String(), "login already exists") {
		t.Errorf("expected error message, got %q", w.Body.String())
	}
}

func TestRegisterEmptyCredentials(t *testing.T) {
	mockService := &MockAuthService{
		registerFunc: func(ctx context.Context, login, password string) error {
			if login == "" || password == "" {
				return errors.New("empty credentials")
			}
			return nil
		},
	}

	handler := NewAuthHandler(mockService)

	tests := []struct {
		name     string
		login    string
		password string
	}{
		{"empty login", "", "password"},
		{"empty password", "login", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := RegisterRequest{
				Login:    tt.login,
				Password: tt.password,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.Register(w, req)

			if w.Code != http.StatusInternalServerError {
				t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
			}
		})
	}
}

func TestRegisterNoBody(t *testing.T) {
	mockService := &MockAuthService{}
	handler := NewAuthHandler(mockService)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader([]byte("")))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLoginNoBody(t *testing.T) {
	mockService := &MockAuthService{}
	handler := NewAuthHandler(mockService)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader([]byte("")))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLoginLargeCredentials(t *testing.T) {
	largeLogin := string(make([]byte, 10000))
	largePassword := string(make([]byte, 10000))

	mockService := &MockAuthService{
		loginFunc: func(ctx context.Context, login, password string) (string, error) {
			return "token", nil
		},
	}

	handler := NewAuthHandler(mockService)

	reqBody := RegisterRequest{
		Login:    largeLogin,
		Password: largePassword,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRegisterMultipleUsers(t *testing.T) {
	registeredUsers := make(map[string]bool)

	mockService := &MockAuthService{
		registerFunc: func(ctx context.Context, login, password string) error {
			if registeredUsers[login] {
				return errors.New("user already exists")
			}
			registeredUsers[login] = true
			return nil
		},
	}

	handler := NewAuthHandler(mockService)

	users := []string{"user1", "user2", "user3"}

	for _, username := range users {
		reqBody := RegisterRequest{
			Login:    username,
			Password: "password",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("failed to register %s: status %d", username, w.Code)
		}
	}

	reqBody := RegisterRequest{
		Login:    "user1",
		Password: "password",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected error for duplicate user, got status %d", w.Code)
	}
}

func TestLoginMultipleAttempts(t *testing.T) {
	mockService := &MockAuthService{
		loginFunc: func(ctx context.Context, login, password string) (string, error) {
			if login == "validuser" && password == "validpass" {
				return "token_valid", nil
			}
			return "", errors.New("invalid credentials")
		},
	}

	handler := NewAuthHandler(mockService)

	attempts := []struct {
		login    string
		password string
		expected bool
	}{
		{"validuser", "validpass", true},
		{"validuser", "wrongpass", false},
		{"wronguser", "validpass", false},
		{"validuser", "validpass", true},
	}

	for i, attempt := range attempts {
		reqBody := RegisterRequest{
			Login:    attempt.login,
			Password: attempt.password,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.Login(w, req)

		if attempt.expected && w.Code != http.StatusOK {
			t.Errorf("attempt %d: expected success (status 200), got %d", i, w.Code)
		}

		if !attempt.expected && w.Code != http.StatusBadRequest {
			t.Errorf("attempt %d: expected failure (status 400), got %d", i, w.Code)
		}
	}
}

func TestNewAuthHandler(t *testing.T) {
	mockService := &MockAuthService{}
	handler := NewAuthHandler(mockService)

	if handler == nil {
		t.Error("expected handler to be created")
	}

	if handler.authService == nil {
		t.Error("expected authService to be set")
	}
}

func TestRegisterResponseStatus(t *testing.T) {
	mockService := &MockAuthService{
		registerFunc: func(ctx context.Context, login, password string) error {
			return nil
		},
	}

	handler := NewAuthHandler(mockService)

	reqBody := RegisterRequest{
		Login:    "user",
		Password: "pass",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if w.Code == http.StatusOK {
		t.Error("register should return 201 Created, not 200 OK")
	}
}

func TestLoginContextUsage(t *testing.T) {
	contextUsed := false

	mockService := &MockAuthService{
		loginFunc: func(ctx context.Context, login, password string) (string, error) {
			contextUsed = ctx != nil
			return "token", nil
		},
	}

	handler := NewAuthHandler(mockService)

	reqBody := RegisterRequest{
		Login:    "user",
		Password: "pass",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if !contextUsed {
		t.Error("expected context to be passed to service")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
