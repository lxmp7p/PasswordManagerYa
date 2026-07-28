package handler

import (
	"encoding/json"
	"net/http"
	"passwordmanager/internal/service"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	authService service.AuthServiceInterface
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) InitRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Post("/register", h.Register)
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "failed to validete request data", http.StatusBadRequest)
		return
	}

	h.authService.Register(r.Context(), req.Login, req.Password)

	w.WriteHeader(http.StatusCreated)
}
