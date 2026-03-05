package scheduler

import (
	"context"
	"log/slog"
	"sync"

	"github.com/robfig/cron/v3"
	"open-maguro/internal/domain"
	"open-maguro/internal/executor"
)

// TaskLoader loads enabled tasks from the database.
type TaskLoader interface {
	ListEnabled(ctx context.Context) ([]domain.AgentTask, error)
}

type Scheduler struct {
	cron     *cron.Cron
	loader   TaskLoader
	executor *executor.Executor
	mu       sync.Mutex
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

func New(loader TaskLoader, exec *executor.Executor) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		cron:     cron.New(),
		loader:   loader,
		executor: exec,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start loads enabled tasks from the DB, registers them with cron, and starts scheduling.
func (s *Scheduler) Start() error {
	if err := s.loadAndRegister(); err != nil {
		return err
	}
	s.cron.Start()
	slog.Info("scheduler started")
	return nil
}

// Reload clears all cron entries and re-registers from the DB.
func (s *Scheduler) Reload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.Info("reloading scheduler")

	for _, entry := range s.cron.Entries() {
		s.cron.Remove(entry.ID)
	}

	if err := s.loadAndRegister(); err != nil {
		slog.Error("failed to reload scheduler", "error", err)
	}
}

// Stop stops the cron scheduler and waits for all running jobs to finish.
func (s *Scheduler) Stop() {
	slog.Info("stopping scheduler")

	stopCtx := s.cron.Stop()
	s.cancel()
	<-stopCtx.Done()
	s.wg.Wait()

	slog.Info("scheduler stopped")
}

func (s *Scheduler) loadAndRegister() error {
	tasks, err := s.loader.ListEnabled(s.ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		task := task
		_, err := s.cron.AddFunc(task.CronExpression, func() {
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				s.executor.Run(s.ctx, task)
			}()
		})
		if err != nil {
			slog.Error("failed to register cron job",
				"task_id", task.ID,
				"task_name", task.Name,
				"cron", task.CronExpression,
				"error", err,
			)
			continue
		}
		slog.Info("registered cron job",
			"task_id", task.ID,
			"task_name", task.Name,
			"cron", task.CronExpression,
		)
	}

	slog.Info("loaded tasks into scheduler", "count", len(tasks))
	return nil
}
