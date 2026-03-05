package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"

	"open-maguro/internal/agent_task"
	"open-maguro/internal/config"
	"open-maguro/internal/database"
	"open-maguro/internal/router"
	"open-maguro/internal/task_execution"
)

func main() {
	cfg, err := env.ParseAs[config.Config]()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	validate := validator.New()

	// Wire up agent_task
	agentTaskRepo := agent_task.NewPostgresRepository(pool)
	agentTaskService := agent_task.NewService(agentTaskRepo)
	agentTaskHandler := agent_task.NewHandler(agentTaskService, validate)

	// Wire up task_execution
	taskExecRepo := task_execution.NewPostgresRepository(pool)
	taskExecService := task_execution.NewService(taskExecRepo)
	taskExecHandler := task_execution.NewHandler(taskExecService)

	r := router.New(agentTaskHandler, taskExecHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down server")
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}()

	slog.Info("server starting", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
