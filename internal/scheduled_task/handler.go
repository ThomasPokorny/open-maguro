package scheduled_task

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service       *Service
	validate      *validator.Validate
	onTaskChanged func()
}

type HandlerOption func(*Handler)

func WithOnTaskChanged(fn func()) HandlerOption {
	return func(h *Handler) {
		h.onTaskChanged = fn
	}
}

func NewHandler(service *Service, validate *validator.Validate, opts ...HandlerOption) *Handler {
	h := &Handler{service: service, validate: validate}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/scheduled-tasks", h.Create)
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

	task, err := h.service.Create(r.Context(), req)
	if err != nil {
		slog.Error("failed to create scheduled task", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create scheduled task")
		return
	}

	writeJSON(w, http.StatusCreated, ToResponse(task))

	if h.onTaskChanged != nil {
		go h.onTaskChanged()
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
