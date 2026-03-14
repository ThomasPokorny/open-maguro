package task_execution

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

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
	r.Delete("/executions", h.Purge)
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

func (h *Handler) Purge(w http.ResponseWriter, r *http.Request) {
	olderThan := r.URL.Query().Get("older_than")
	if olderThan == "" {
		writeError(w, http.StatusBadRequest, "older_than query parameter is required (RFC3339 timestamp or duration like '30d', '24h')")
		return
	}

	var before time.Time
	// Try parsing as duration shorthand (e.g. "30d", "24h", "7d")
	if parsed, err := parseDuration(olderThan); err == nil {
		before = time.Now().Add(-parsed)
	} else if parsed, err := time.Parse(time.RFC3339, olderThan); err == nil {
		before = parsed
	} else {
		writeError(w, http.StatusBadRequest, "invalid older_than value: use RFC3339 timestamp or duration (e.g. '30d', '24h')")
		return
	}

	count, err := h.service.DeleteOlderThan(r.Context(), before)
	if err != nil {
		slog.Error("failed to purge executions", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to purge executions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"deleted": count})
}

// parseDuration handles Go-style durations plus "d" suffix for days.
func parseDuration(s string) (time.Duration, error) {
	if len(s) > 1 && s[len(s)-1] == 'd' {
		// Parse "30d" as 30 * 24h
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err == nil && days > 0 {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}
	return time.ParseDuration(s)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
