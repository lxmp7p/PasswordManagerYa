package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Login_Success(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			assert.Equal(
				t,
				"/auth/login",
				r.URL.Path,
			)

			assert.Equal(
				t,
				http.MethodPost,
				r.Method,
			)

			assert.Equal(
				t,
				"application/json",
				r.Header.Get("Content-Type"),
			)

			var body LoginRequest

			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)

			assert.Equal(t, "admin", body.Login)
			assert.Equal(t, "123", body.Password)

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			json.NewEncoder(w).Encode(
				LoginResponse{
					Token: "test-token",
				},
			)
		}),
	)

	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	err := client.Login(
		"admin",
		"123",
	)

	require.NoError(t, err)

	assert.Equal(
		t,
		"test-token",
		client.Token,
	)
}

func TestClient_Register_Success(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			assert.Equal(
				t,
				"/auth/register",
				r.URL.Path,
			)

			assert.Equal(
				t,
				http.StatusCreated,
				http.StatusCreated,
			)

			w.WriteHeader(
				http.StatusCreated,
			)
		}),
	)

	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	err := client.Register(
		"user",
		"password",
	)

	require.NoError(t, err)
}

func TestClient_Register_Error(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.WriteHeader(
				http.StatusBadRequest,
			)

			w.Write([]byte(
				"invalid user",
			))
		}),
	)

	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	err := client.Register(
		"user",
		"password",
	)

	assert.Error(t, err)

	assert.Contains(
		t,
		err.Error(),
		"invalid user",
	)
}

func TestClient_CreateVault_Success(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			assert.Equal(
				t,
				"/vault",
				r.URL.Path,
			)

			assert.Equal(
				t,
				"Bearer token",
				r.Header.Get("Authorization"),
			)

			var body VaultCreate

			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)

			assert.Equal(
				t,
				"github",
				body.Title,
			)

			w.WriteHeader(
				http.StatusCreated,
			)
		}),
	)

	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		Token:   "token",
		Client:  server.Client(),
	}

	err := client.CreateVault(
		VaultCreate{
			Type:  "password",
			Title: "github",
			Data: map[string]any{
				"password": "secret",
			},
		},
	)

	require.NoError(t, err)
}

func TestClient_ListVault_Success(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			assert.Equal(
				t,
				"Bearer token",
				r.Header.Get("Authorization"),
			)

			json.NewEncoder(w).Encode(
				[]VaultItem{
					{
						ID:    1,
						Title: "github",
						Type:  "password",
					},
				},
			)
		}),
	)

	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		Token:   "token",
		Client:  server.Client(),
	}

	items, err := client.ListVault()

	require.NoError(t, err)

	require.Len(
		t,
		items,
		1,
	)

	assert.Equal(
		t,
		int64(1),
		items[0].ID,
	)

	assert.Equal(
		t,
		"github",
		items[0].Title,
	)
}

func TestClient_ListVault_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.WriteHeader(
				http.StatusUnauthorized,
			)
		}),
	)

	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	_, err := client.ListVault()

	assert.Error(t, err)
}

func TestClient_GetVaultByID_Success(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			assert.Equal(
				t,
				"/vault/10",
				r.URL.Path,
			)

			json.NewEncoder(w).Encode(
				VaultItem{
					ID:    10,
					Title: "google",
				},
			)
		}),
	)

	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		Token:   "token",
		Client:  server.Client(),
	}

	item, err := client.GetVaultByID(10)

	require.NoError(t, err)

	assert.Equal(
		t,
		int64(10),
		item.ID,
	)

	assert.Equal(
		t,
		"google",
		item.Title,
	)
}

func TestClient_GetVaultByID_Error(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.WriteHeader(
				http.StatusNotFound,
			)
		}),
	)

	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	_, err := client.GetVaultByID(999)

	assert.Error(t, err)
}

func TestClient_Login_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Write([]byte(
				"broken json",
			))
		}),
	)

	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	err := client.Login(
		"user",
		"pass",
	)

	assert.Error(t, err)
}

func TestClient_Register_ServerUnavailable(t *testing.T) {
	client := &Client{
		BaseURL: "http://localhost:1",
		Client:  http.DefaultClient,
	}

	err := client.Register(
		"user",
		"pass",
	)

	assert.True(
		t,
		errors.Is(err, err),
	)
}
