package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	dbpkg "open-maguro/db"
	"open-maguro/internal/agent_task"
	"open-maguro/internal/config"
	"open-maguro/internal/database"
	"open-maguro/internal/executor"
	"open-maguro/internal/mcp_config"
	"open-maguro/internal/router"
	"open-maguro/internal/scheduled_task"
	"open-maguro/internal/scheduler"
	"open-maguro/internal/skill"
	"open-maguro/internal/task_execution"
)

func main() {
	_ = godotenv.Load() // optional: loads .env if present

	cfg, err := env.ParseAs[config.Config]()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run migrations
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	validate := validator.New()

	// Wire up repositories
	agentTaskRepo := agent_task.NewPostgresRepository(pool)
	taskExecRepo := task_execution.NewPostgresRepository(pool)
	skillRepo := skill.NewPostgresRepository(pool)

	// Wire up executor and scheduler
	exec := executor.New(taskExecRepo, skillRepo, cfg.MCPConfigPath, cfg.AllowedTools)
	sched := scheduler.New(agentTaskRepo, agentTaskRepo, exec)

	// Wire up agent_task (with scheduler reload callback)
	agentTaskService := agent_task.NewService(agentTaskRepo)
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

	r := router.New(agentTaskHandler, taskExecHandler, scheduledTaskHandler, mcpConfigHandler, skillHandler)

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
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}()

	slog.Info("server starting", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func runMigrations(databaseURL string) error {
	goose.SetBaseFS(dbpkg.Migrations)

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, "migrations")
}
