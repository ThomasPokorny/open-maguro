package skill

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
	r.Route("/skills", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.Get)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})

	r.Get("/agent-tasks/{id}/skills", h.ListByAgentTask)
	r.Post("/agent-tasks/{id}/skills/{skillId}", h.AddAgentSkill)
	r.Delete("/agent-tasks/{id}/skills/{skillId}", h.RemoveAgentSkill)
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

	s, err := h.service.Create(r.Context(), req)
	if err != nil {
		slog.Error("failed to create skill", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create skill")
		return
	}

	writeJSON(w, http.StatusCreated, ToResponse(s))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	skills, err := h.service.List(r.Context())
	if err != nil {
		slog.Error("failed to list skills", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list skills")
		return
	}

	writeJSON(w, http.StatusOK, ToResponseList(skills))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	s, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}

	writeJSON(w, http.StatusOK, ToResponse(s))
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

	s, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}

	writeJSON(w, http.StatusOK, ToResponse(s))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		slog.Error("failed to delete skill", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete skill")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListByAgentTask(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	skills, err := h.service.ListByAgentTaskID(r.Context(), id)
	if err != nil {
		slog.Error("failed to list skills for agent task", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list skills")
		return
	}

	writeJSON(w, http.StatusOK, ToResponseList(skills))
}

func (h *Handler) AddAgentSkill(w http.ResponseWriter, r *http.Request) {
	agentTaskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid agent task id")
		return
	}

	skillID, err := uuid.Parse(chi.URLParam(r, "skillId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid skill id")
		return
	}

	if err := h.service.AddAgentSkill(r.Context(), agentTaskID, skillID); err != nil {
		slog.Error("failed to add agent skill", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to add skill to agent")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveAgentSkill(w http.ResponseWriter, r *http.Request) {
	agentTaskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid agent task id")
		return
	}

	skillID, err := uuid.Parse(chi.URLParam(r, "skillId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid skill id")
		return
	}

	if err := h.service.RemoveAgentSkill(r.Context(), agentTaskID, skillID); err != nil {
		slog.Error("failed to remove agent skill", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to remove skill from agent")
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
