package router

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"open-maguro/internal/agent_task"
	"open-maguro/internal/kanban"
	"open-maguro/internal/maguro_chat"
	"open-maguro/internal/mcp_config"
	"open-maguro/internal/scheduled_task"
	"open-maguro/internal/skill"
	"open-maguro/internal/task_execution"
	"open-maguro/internal/team"
)

// Option configures the router.
type Option func(chi.Router)

// WithStaticFS serves an embedded filesystem as the SPA frontend.
func WithStaticFS(staticFS fs.FS) Option {
	return func(r chi.Router) {
		fileServer := http.FileServer(http.FS(staticFS))
		indexHTML, _ := fs.ReadFile(staticFS, "index.html")
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			// Serve actual files (JS, CSS, images) directly
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path != "" && strings.Contains(path, ".") {
				fileServer.ServeHTTP(w, r)
				return
			}
			// SPA fallback: serve index.html content directly
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(indexHTML)
		})
	}
}

func New(agentTaskHandler *agent_task.Handler, taskExecHandler *task_execution.Handler, scheduledTaskHandler *scheduled_task.Handler, mcpConfigHandler *mcp_config.Handler, skillHandler *skill.Handler, kanbanHandler *kanban.Handler, teamHandler *team.Handler, maguroChatHandler *maguro_chat.Handler, opts ...Option) chi.Router {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
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
		scheduledTaskHandler.RegisterRoutes(r)
		mcpConfigHandler.RegisterRoutes(r)
		skillHandler.RegisterRoutes(r)
		kanbanHandler.RegisterRoutes(r)
		teamHandler.RegisterRoutes(r)
		maguroChatHandler.RegisterRoutes(r)
	})

	for _, opt := range opts {
		opt(r)
	}

	return r
}
