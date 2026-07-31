package handler

import (
	"encoding/json"
	"net/http"
	"passwordmanager/internal/dto"
	"passwordmanager/internal/handler/middlewares"
	"passwordmanager/internal/model"
	"passwordmanager/internal/service"

	"github.com/go-chi/chi/v5"
)

type VaultHandler struct {
	tokenService service.TokenManagerInterface
	vaultService service.VaultServiceInterface
}

func NewVaultHandler(
	vaultService service.VaultServiceInterface,
	tokenService service.TokenManagerInterface) *VaultHandler {
	return &VaultHandler{
		vaultService: vaultService,
		tokenService: tokenService,
	}
}

func (h *VaultHandler) InitRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares.JWT(h.tokenService))

		r.Route("/vault", func(r chi.Router) {
			r.Post("/", h.Create)
		})
	})
}

func (h *VaultHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.VaultCreateInDto

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "failed to validate request data", http.StatusBadRequest)
		return
	}

	userID, ok := middlewares.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	item := service.VaultCreate{
		Type:     model.ItemType(req.Type),
		Title:    req.Title,
		Data:     req.Data,
		Metadata: req.Metadata,
	}

	err = h.vaultService.Create(r.Context(), userID, item)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
