package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"passwordmanager/internal/service"

	"github.com/golang-jwt/jwt/v5"
)

type MockTokenManager struct {
	parseFunc    func(token string) (*service.Claims, error)
	generateFunc func(userID int64, subject string) (string, error)
}

func (m *MockTokenManager) Parse(token string) (*service.Claims, error) {
	if m.parseFunc != nil {
		return m.parseFunc(token)
	}
	return nil, errors.New("no parse function configured")
}

func (m *MockTokenManager) Generate(userID int64, subject string) (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc(userID, subject)
	}
	return "", errors.New("not implemented in mock")
}

func TestJWTMissingAuthHeader(t *testing.T) {
	mockManager := &MockTokenManager{}
	middleware := JWT(mockManager)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/vault", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if !contains(w.Body.String(), "missing authorization header") {
		t.Errorf("expected 'missing authorization header', got %q", w.Body.String())
	}
}

func TestJWTInvalidAuthHeaderFormat(t *testing.T) {
	mockManager := &MockTokenManager{}
	middleware := JWT(mockManager)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []string{
		"Basic xyz",
		"Bearer",
		"BearerToken",
		"bearer token",
	}

	for _, authHeader := range tests {
		t.Run("invalid_format_"+authHeader, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/vault", nil)
			req.Header.Set("Authorization", authHeader)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d for %q", http.StatusBadRequest, w.Code, authHeader)
			}
		})
	}
}

func TestJWTInvalidToken(t *testing.T) {
	mockManager := &MockTokenManager{
		parseFunc: func(token string) (*service.Claims, error) {
			return nil, errors.New("invalid token")
		},
	}
	middleware := JWT(mockManager)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/vault", nil)
	req.Header.Set("Authorization", "Bearer invalid_token_xyz")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestJWTValidToken(t *testing.T) {
	mockManager := &MockTokenManager{
		parseFunc: func(token string) (*service.Claims, error) {
			if token == "valid_token" {
				return &service.Claims{
					Login: "testuser",
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "123",
					},
				}, nil
			}
			return nil, errors.New("invalid token")
		},
	}
	middleware := JWT(mockManager)

	handlerCalled := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Error("expected userID in context")
			return
		}
		if userID != 123 {
			t.Errorf("expected userID 123, got %d", userID)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/vault", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("expected next handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestJWTInvalidUserID(t *testing.T) {
	mockManager := &MockTokenManager{
		parseFunc: func(token string) (*service.Claims, error) {
			return &service.Claims{
				Login: "testuser",
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "not_a_number",
				},
			}, nil
		},
	}
	middleware := JWT(mockManager)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/vault", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	if !contains(w.Body.String(), "invalid user id") {
		t.Errorf("expected 'invalid user id', got %q", w.Body.String())
	}
}

func TestUserIDFromContext(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		expectedID    int64
		expectedFound bool
	}{
		{
			name:          "valid userID",
			ctx:           context.WithValue(context.Background(), userIDKey, int64(42)),
			expectedID:    42,
			expectedFound: true,
		},
		{
			name:          "missing userID",
			ctx:           context.Background(),
			expectedID:    0,
			expectedFound: false,
		},
		{
			name:          "wrong type in context",
			ctx:           context.WithValue(context.Background(), userIDKey, "not_an_int"),
			expectedID:    0,
			expectedFound: false,
		},
		{
			name:          "zero userID",
			ctx:           context.WithValue(context.Background(), userIDKey, int64(0)),
			expectedID:    0,
			expectedFound: true,
		},
		{
			name:          "large userID",
			ctx:           context.WithValue(context.Background(), userIDKey, int64(9223372036854775807)),
			expectedID:    9223372036854775807,
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, ok := UserIDFromContext(tt.ctx)

			if ok != tt.expectedFound {
				t.Errorf("expected found=%v, got %v", tt.expectedFound, ok)
			}

			if userID != tt.expectedID {
				t.Errorf("expected userID=%d, got %d", tt.expectedID, userID)
			}
		})
	}
}

func TestJWTLargeUserID(t *testing.T) {
	largeID := int64(9223372036854775807)
	mockManager := &MockTokenManager{
		parseFunc: func(token string) (*service.Claims, error) {
			return &service.Claims{
				Login: "testuser",
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "9223372036854775807",
				},
			}, nil
		},
	}
	middleware := JWT(mockManager)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Error("expected userID in context")
			return
		}
		if userID != largeID {
			t.Errorf("expected userID %d, got %d", largeID, userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/vault", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestJWTContextPropagation(t *testing.T) {
	mockManager := &MockTokenManager{
		parseFunc: func(token string) (*service.Claims, error) {
			return &service.Claims{
				Login: "testuser",
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "999",
				},
			}, nil
		},
	}
	middleware := JWT(mockManager)

	contextChecked := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextChecked = true

		_, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Error("userID not found in context")
			return
		}

		if r.Context() == nil {
			t.Error("context is nil")
			return
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/vault", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !contextChecked {
		t.Error("context was not checked in handler")
	}
}

func TestJWTMultipleRequests(t *testing.T) {
	mockManager := &MockTokenManager{
		parseFunc: func(token string) (*service.Claims, error) {
			subject := token[5:]
			return &service.Claims{
				Login: "testuser",
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: subject,
				},
			}, nil
		},
	}
	middleware := JWT(mockManager)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	requests := []struct {
		token      string
		expectedID int64
	}{
		{"user_1", 1},
		{"user_42", 42},
		{"user_100", 100},
	}

	for _, req := range requests {
		httpReq := httptest.NewRequest("GET", "/vault", nil)
		httpReq.Header.Set("Authorization", "Bearer "+req.token)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, httpReq)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	}
}

func TestJWTWhitespaceInToken(t *testing.T) {
	mockManager := &MockTokenManager{
		parseFunc: func(token string) (*service.Claims, error) {
			return nil, errors.New("invalid token")
		},
	}
	middleware := JWT(mockManager)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name      string
		authHead  string
		expectErr bool
	}{
		{"token with spaces", "Bearer token with spaces", true},
		{"normal token", "Bearer validtoken", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/vault", nil)
			req.Header.Set("Authorization", test.authHead)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if test.expectErr && w.Code < 400 {
				t.Errorf("expected error status, got %d", w.Code)
			}
		})
	}
}

func TestUserIDKeyConstant(t *testing.T) {
	if userIDKey == "" {
		t.Error("UserIDKey should not be empty")
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, int64(123))

	userID, ok := UserIDFromContext(ctx)
	if !ok || userID != 123 {
		t.Error("UserIDKey constant is not working correctly")
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
