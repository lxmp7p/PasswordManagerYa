package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"passwordmanager/internal/dto"
	"passwordmanager/internal/handler/middlewares"
	"passwordmanager/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TokenManager struct {
	mock.Mock
}

func (m *TokenManager) Generate(userID int64, login string) (string, error) {
	args := m.Called(userID, login)

	return args.String(0), args.Error(1)
}

func (m *TokenManager) Parse(token string) (*service.Claims, error) {
	args := m.Called(token)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*service.Claims), args.Error(1)
}

type VaultService struct {
	mock.Mock
}

func (m *VaultService) Create(
	ctx context.Context,
	userID int64,
	item service.VaultCreate,
) (int64, error) {
	args := m.Called(ctx, userID, item)

	return args.Get(0).(int64), args.Error(1)
}

func (m *VaultService) Get(
	ctx context.Context,
	id int64,
	userID int64,
) (dto.VaultResponseGet, error) {

	args := m.Called(ctx, id, userID)

	return args.Get(0).(dto.VaultResponseGet), args.Error(1)
}

func contextWithUserID(
	ctx context.Context,
	userID int64,
) context.Context {
	return context.WithValue(
		ctx,
		middlewares.UserIDKey,
		userID,
	)
}

func (m *VaultService) List(
	ctx context.Context,
	userID int64,
) ([]dto.VaultResponseGet, error) {
	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]dto.VaultResponseGet), args.Error(1)
}

func setupVaultRouter(vaultService *VaultService) chi.Router {
	h := NewVaultHandler(
		vaultService,
		nil,
	)

	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctx := contextWithUserID(
				r.Context(),
				123,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	})

	r.Post("/vault/", h.Create)
	r.Get("/vault/", h.List)
	r.Get("/vault/{id}", h.Get)

	return r
}
func TestVaultHandler_Create_Success(t *testing.T) {
	vaultService := new(VaultService)

	vaultService.
		On(
			"Create",
			mock.Anything,
			int64(123),
			mock.Anything,
		).
		Return(int64(10), nil)

	router := setupVaultRouter(vaultService)

	body := `
{
	"type":"password",
	"title":"github",
	"data":{
		"password":"secret"
	},
	"metadata":{
		"source":"github"
	}
}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/vault/",
		bytes.NewBufferString(body),
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusCreated,
		rec.Code,
	)

	assert.JSONEq(
		t,
		"10",
		rec.Body.String(),
	)

	vaultService.AssertExpectations(t)
}

func TestVaultHandler_Get_Success(t *testing.T) {
	vaultService := new(VaultService)

	vaultService.
		On(
			"Get",
			mock.Anything,
			int64(1),
			int64(123),
		).
		Return(
			dto.VaultResponseGet{
				ID:    1,
				Title: "github",
			},
			nil,
		)

	router := setupVaultRouter(vaultService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/vault/1",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusOK,
		rec.Code,
	)

	assert.JSONEq(
		t,
		`{
		"id":1,
		"title":"github",
		"type":"",
		"data":null,
		"metadata":null,
		"created_at":"0001-01-01T00:00:00Z",
		"updated_at":"0001-01-01T00:00:00Z"
	}`,
		rec.Body.String(),
	)

	vaultService.AssertExpectations(t)
}

func TestVaultHandler_Get_InvalidID(t *testing.T) {
	vaultService := new(VaultService)

	router := setupVaultRouter(vaultService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/vault/test",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusBadRequest,
		rec.Code,
	)

	vaultService.AssertNotCalled(
		t,
		"Get",
	)
}
func TestVaultHandler_Get_ServiceError(t *testing.T) {
	vaultService := new(VaultService)

	vaultService.
		On(
			"Get",
			mock.Anything,
			int64(1),
			int64(123),
		).
		Return(
			dto.VaultResponseGet{},
			errors.New("not found"),
		)

	router := setupVaultRouter(vaultService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/vault/1",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusInternalServerError,
		rec.Code,
	)

	vaultService.AssertExpectations(t)
}

func TestVaultHandler_List_Success(t *testing.T) {
	vaultService := new(VaultService)

	vaultService.
		On(
			"List",
			mock.Anything,
			int64(123),
		).
		Return(
			[]dto.VaultResponseGet{
				{
					ID:    1,
					Title: "github",
				},
				{
					ID:    2,
					Title: "google",
				},
			},
			nil,
		)

	router := setupVaultRouter(vaultService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/vault/",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusOK,
		rec.Code,
	)

	assert.JSONEq(
		t,
		`[
		{
			"id":1,
			"title":"github",
			"type":"",
			"data":null,
			"metadata":null,
			"created_at":"0001-01-01T00:00:00Z",
			"updated_at":"0001-01-01T00:00:00Z"
		},
		{
			"id":2,
			"title":"google",
			"type":"",
			"data":null,
			"metadata":null,
			"created_at":"0001-01-01T00:00:00Z",
			"updated_at":"0001-01-01T00:00:00Z"
		}
	]`,
		rec.Body.String(),
	)

	vaultService.AssertExpectations(t)
}
func TestVaultHandler_List_ServiceError(t *testing.T) {
	vaultService := new(VaultService)

	vaultService.
		On(
			"List",
			mock.Anything,
			int64(123),
		).
		Return(
			nil,
			errors.New("database error"),
		)

	router := setupVaultRouter(vaultService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/vault/",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusInternalServerError,
		rec.Code,
	)

	vaultService.AssertExpectations(t)
}
