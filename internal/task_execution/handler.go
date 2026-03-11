package task_execution

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/agent-tasks/{taskId}/executions", h.ListByAgentTask)
	r.Get("/executions", h.List)
	r.Get("/executions/{id}", h.Get)
}

func (h *Handler) ListByAgentTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	executions, err := h.service.ListByAgentTaskID(r.Context(), taskID)
	if err != nil {
		slog.Error("failed to list executions", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list executions")
		return
	}

	writeJSON(w, http.StatusOK, ToResponseList(executions))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	executions, err := h.service.List(r.Context())
	if err != nil {
		slog.Error("failed to list executions", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list executions")
		return
	}
	writeJSON(w, http.StatusOK, ToResponseList(executions))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	execution, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "execution not found")
		return
	}

	writeJSON(w, http.StatusOK, ToResponse(execution))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
