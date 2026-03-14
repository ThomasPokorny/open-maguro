package maguro_chat

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

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
	r.Post("/chat", h.Chat)
	r.Post("/chat/reset", h.Reset)
}

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "message is required")
		return
	}

	slog.Info("maguro chat request", "message_length", len(req.Message))

	// Use a generous timeout — Claude calls can take minutes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	reply, err := h.service.Chat(ctx, req.Message)
	if err != nil {
		slog.Error("maguro chat failed", "error", err)
		writeError(w, http.StatusInternalServerError, "chat failed: "+err.Error())
		return
	}

	sessionID := h.service.SessionID()
	slog.Info("maguro chat response", "reply_length", len(reply), "session_id", sessionID)
	writeJSON(w, http.StatusOK, ChatResponse{Reply: reply, SessionID: sessionID})
}

func (h *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	h.service.ResetSession()
	slog.Info("maguro chat session reset")
	writeJSON(w, http.StatusOK, map[string]string{"status": "session reset"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
