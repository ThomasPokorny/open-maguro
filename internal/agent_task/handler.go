package agent_task

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

// AgentRunner can execute an agent task.
type AgentRunner interface {
	Run(ctx context.Context, task domain.AgentTask, onComplete func())
}

type Handler struct {
	service       *Service
	validate      *validator.Validate
	onTaskChanged func()
	runner        AgentRunner
}

type HandlerOption func(*Handler)

func WithOnTaskChanged(fn func()) HandlerOption {
	return func(h *Handler) {
		h.onTaskChanged = fn
	}
}

func WithRunner(runner AgentRunner) HandlerOption {
	return func(h *Handler) {
		h.runner = runner
	}
}

func NewHandler(service *Service, validate *validator.Validate, opts ...HandlerOption) *Handler {
	h := &Handler{service: service, validate: validate}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) notifyTaskChanged() {
	if h.onTaskChanged != nil {
		go h.onTaskChanged()
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/agent-tasks", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.Get)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Post("/{id}/run", h.Run)
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

	task, err := h.service.Create(r.Context(), req)
	if err != nil {
		slog.Error("failed to create agent task", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create agent task")
		return
	}

	writeJSON(w, http.StatusCreated, ToResponse(task))
	h.notifyTaskChanged()
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.List(r.Context())
	if err != nil {
		slog.Error("failed to list agent tasks", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list agent tasks")
		return
	}

	writeJSON(w, http.StatusOK, ToResponseList(tasks))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	task, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "agent task not found")
		return
	}

	writeJSON(w, http.StatusOK, ToResponse(task))
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

	task, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "circular chain detected") {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusNotFound, "agent task not found")
		return
	}

	writeJSON(w, http.StatusOK, ToResponse(task))
	h.notifyTaskChanged()
}

func (h *Handler) Run(w http.ResponseWriter, r *http.Request) {
	if h.runner == nil {
		writeError(w, http.StatusInternalServerError, "agent runner not configured")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	task, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "agent task not found")
		return
	}

	go h.runner.Run(context.Background(), *task, nil)

	slog.Info("agent task run triggered", "task_id", task.ID, "task_name", task.Name)
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		slog.Error("failed to delete agent task", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete agent task")
		return
	}

	w.WriteHeader(http.StatusNoContent)
	h.notifyTaskChanged()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
