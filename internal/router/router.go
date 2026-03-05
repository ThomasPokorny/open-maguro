package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"open-maguro/internal/agent_task"
	"open-maguro/internal/task_execution"
)

func New(agentTaskHandler *agent_task.Handler, taskExecHandler *task_execution.Handler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		agentTaskHandler.RegisterRoutes(r)
		taskExecHandler.RegisterRoutes(r)
	})

	return r
}
