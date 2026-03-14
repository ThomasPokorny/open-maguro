package team

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(service *Service, validate *validator.Validate) *Handler {
	return &Handler{service: service, validate: validate}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/teams", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.Get)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	t, err := h.service.Create(r.Context(), req)
	if err != nil {
		slog.Error("failed to create team", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create team")
		return
	}

	writeJSON(w, http.StatusCreated, ToResponse(t))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	teams, err := h.service.List(r.Context())
	if err != nil {
		slog.Error("failed to list teams", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list teams")
		return
	}

	writeJSON(w, http.StatusOK, ToResponseList(teams))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	t, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "team not found")
		return
	}

	writeJSON(w, http.StatusOK, ToResponse(t))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	t, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		writeError(w, http.StatusNotFound, "team not found")
		return
	}

	writeJSON(w, http.StatusOK, ToResponse(t))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		slog.Error("failed to delete team", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete team")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
