package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"passwordmanager/internal/dto"
	"passwordmanager/internal/handler/middlewares"
	"passwordmanager/internal/model"
	"passwordmanager/internal/service"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type VaultHandler struct {
	logger       *slog.Logger
	tokenService service.TokenManagerInterface
	vaultService service.VaultServiceInterface
}

func NewVaultHandler(
	logger *slog.Logger,
	vaultService service.VaultServiceInterface,
	tokenService service.TokenManagerInterface) *VaultHandler {
	return &VaultHandler{
		logger:       logger,
		vaultService: vaultService,
		tokenService: tokenService,
	}
}

func (h *VaultHandler) InitRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares.JWT(h.tokenService))

		r.Route("/vault", func(r chi.Router) {
			r.Get("/", h.List)
			r.Get("/{id}", h.Get)
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

	id, err := h.vaultService.Create(r.Context(), userID, item)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		h.logger.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(id)
}

func (h *VaultHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	userID, ok := middlewares.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vault, err := h.vaultService.Get(r.Context(), id, userID)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		h.logger.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vault)
}

func (h *VaultHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vaults, err := h.vaultService.List(r.Context(), userID)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		h.logger.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vaults)
}
