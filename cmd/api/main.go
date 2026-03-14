package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	dbpkg "open-maguro/db"
	"open-maguro/internal/agent_task"
	"open-maguro/internal/config"
	"open-maguro/internal/crypto"
	"open-maguro/internal/database"
	"open-maguro/internal/executor"
	"open-maguro/internal/kanban"
	kanbanexec "open-maguro/internal/kanban_executor"
	"open-maguro/internal/mcp_config"
	"open-maguro/internal/router"
	"open-maguro/internal/scheduled_task"
	"open-maguro/internal/scheduler"
	"open-maguro/internal/skill"
	"open-maguro/internal/task_execution"
	"open-maguro/internal/team"
	dashboard "open-maguro/maguro-dashboard"
)

const banner = `
▛▌▛▌█▌▛▌▄▖▛▛▌▀▌▛▌▌▌▛▘▛▌
▙▌▙▌▙▖▌▌  ▌▌▌█▌▙▌▙▌▌ ▙▌
  ▌            ▄▌
`

func main() {
	_ = godotenv.Load() // optional: loads .env if present

	cfg, err := env.ParseAs[config.Config]()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Expand ~ in database path
	dbPath := cfg.DatabaseURL
	if strings.HasPrefix(dbPath, "~/") {
		home, _ := os.UserHomeDir()
		dbPath = filepath.Join(home, dbPath[2:])
	}
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		slog.Error("failed to create database directory", "error", err)
		os.Exit(1)
	}

	// Run migrations
	if err := runMigrations(dbPath); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	db, err := database.Open(dbPath)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	validate := validator.New()

	// Expand ~ in workspace root
	workspaceRoot := cfg.WorkspaceRoot
	if strings.HasPrefix(workspaceRoot, "~/") {
		home, _ := os.UserHomeDir()
		workspaceRoot = filepath.Join(home, workspaceRoot[2:])
	}
	if workspaceRoot != "" {
		if err := os.MkdirAll(workspaceRoot, 0755); err != nil {
			slog.Error("failed to create workspace root", "path", workspaceRoot, "error", err)
			os.Exit(1)
		}
		slog.Info("workspace root ready", "path", workspaceRoot)
	}

	// Resolve secret key for skill encryption
	var secretKey []byte
	if cfg.SecretKey != "" {
		secretKey, err = hex.DecodeString(cfg.SecretKey)
		if err != nil || len(secretKey) != 32 {
			slog.Error("MAGURO_SECRET_KEY must be a 64-character hex string (32 bytes)")
			os.Exit(1)
		}
	} else {
		home, _ := os.UserHomeDir()
		keyPath := filepath.Join(home, ".maguro", ".secret_key")
		secretKey, err = crypto.LoadOrGenerateKey(keyPath)
		if err != nil {
			slog.Error("failed to load or generate secret key", "error", err)
			os.Exit(1)
		}
		slog.Info("using auto-generated secret key", "path", keyPath)
	}

	// Wire up repositories
	agentTaskRepo := agent_task.NewPostgresRepository(db)
	taskExecRepo := task_execution.NewPostgresRepository(db)
	skillRepo := skill.NewPostgresRepository(db, secretKey)

	// Wire up executor and scheduler
	exec := executor.New(taskExecRepo, skillRepo, agentTaskRepo, cfg.MCPConfigPath, cfg.AllowedTools, workspaceRoot)
	sched := scheduler.New(agentTaskRepo, agentTaskRepo, taskExecRepo, exec, cfg.ExecutionRetentionDays)

	// Wire up agent_task (with scheduler reload callback)
	agentTaskService := agent_task.NewService(agentTaskRepo, workspaceRoot)
	agentTaskHandler := agent_task.NewHandler(agentTaskService, validate,
		agent_task.WithOnTaskChanged(sched.Reload),
		agent_task.WithRunner(exec),
	)

	// Wire up task_execution
	taskExecService := task_execution.NewService(taskExecRepo)
	taskExecHandler := task_execution.NewHandler(taskExecService)

	// Wire up scheduled_task (with scheduler reload callback)
	scheduledTaskService := scheduled_task.NewService(agentTaskRepo)
	scheduledTaskHandler := scheduled_task.NewHandler(scheduledTaskService, validate,
		scheduled_task.WithOnTaskChanged(sched.Reload),
	)

	// Wire up mcp_config
	mcpConfigService := mcp_config.NewService(cfg.MCPConfigPath)
	mcpConfigHandler := mcp_config.NewHandler(mcpConfigService, validate)

	// Wire up skill
	skillService := skill.NewService(skillRepo)
	skillHandler := skill.NewHandler(skillService, validate)

	// Wire up teams
	teamRepo := team.NewPostgresRepository(db)
	teamService := team.NewService(teamRepo)
	teamHandler := team.NewHandler(teamService, validate)

	// Wire up kanban
	kanbanRepo := kanban.NewPostgresRepository(db)
	kanbanService := kanban.NewService(kanbanRepo)
	kanbanExec := kanbanexec.New(kanbanRepo, agentTaskRepo, taskExecRepo, exec)
	if err := kanbanExec.LoadPending(); err != nil {
		slog.Error("failed to load pending kanban tasks", "error", err)
	}
	kanbanHandler := kanban.NewHandler(kanbanService, validate,
		kanban.WithOnTaskCreated(kanbanExec.Enqueue),
	)

	// Serve embedded dashboard
	staticFS, err := fs.Sub(dashboard.Static, "dist")
	if err != nil {
		slog.Error("failed to load embedded dashboard", "error", err)
		os.Exit(1)
	}

	r := router.New(agentTaskHandler, taskExecHandler, scheduledTaskHandler, mcpConfigHandler, skillHandler, kanbanHandler, teamHandler,
		router.WithStaticFS(staticFS),
	)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start scheduler
	if err := sched.Start(); err != nil {
		slog.Error("failed to start scheduler", "error", err)
		os.Exit(1)
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down")
		sched.Stop()
		kanbanExec.Stop()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}()

	fmt.Print(banner)
	fmt.Println("OpenMaguro🐟 v0.2.0 — swim upstream, think downstream.")
	fmt.Printf("🎏Dashboard: http://localhost:%s\n\n", cfg.Port)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func runMigrations(dbPath string) error {
	goose.SetBaseFS(dbpkg.Migrations)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Enable foreign keys for migration
	db.Exec("PRAGMA foreign_keys=ON")

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	return goose.Up(db, "migrations")
}
