package mcp_config

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(service *Service, validate *validator.Validate) *Handler {
	return &Handler{service: service, validate: validate}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/mcp-servers", h.List)
	r.Post("/mcp-servers", h.Add)
	r.Delete("/mcp-servers/{name}", h.Remove)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	servers, err := h.service.List()
	if err != nil {
		slog.Error("failed to list MCP servers", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list MCP servers")
		return
	}
	writeJSON(w, http.StatusOK, servers)
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	var req AddServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if err := h.service.Add(req); err != nil {
		slog.Error("failed to add MCP server", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to add MCP server")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok", "name": req.Name})
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := h.service.Remove(name); err != nil {
		if err == ErrServerNotFound {
			writeError(w, http.StatusNotFound, "MCP server not found")
			return
		}
		slog.Error("failed to remove MCP server", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to remove MCP server")
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
