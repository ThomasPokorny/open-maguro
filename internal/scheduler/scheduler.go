package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"open-maguro/internal/domain"
	"open-maguro/internal/executor"
)

// TaskLoader loads tasks from the database.
type TaskLoader interface {
	ListEnabled(ctx context.Context) ([]domain.AgentTask, error)
	ListPendingScheduled(ctx context.Context) ([]domain.AgentTask, error)
}

// TaskDeleter deletes a task by ID (used for auto-deleting one-time tasks).
type TaskDeleter interface {
	Delete(ctx context.Context, id uuid.UUID) error
}

// ExecutionChecker checks execution history (used by heartbeat).
type ExecutionChecker interface {
	GetLatestByAgentTaskID(ctx context.Context, agentTaskID uuid.UUID) (*domain.TaskExecution, error)
	MarkStaleExecutionsFailed(ctx context.Context, staleBefore time.Time) (int, error)
}

type Scheduler struct {
	cron      *cron.Cron
	loader    TaskLoader
	deleter   TaskDeleter
	execCheck ExecutionChecker
	executor  *executor.Executor
	mu        sync.Mutex
	wg        sync.WaitGroup
	timers    map[uuid.UUID]*time.Timer
	ctx       context.Context
	cancel    context.CancelFunc
}

func New(loader TaskLoader, deleter TaskDeleter, execCheck ExecutionChecker, exec *executor.Executor) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		cron:      cron.New(),
		loader:    loader,
		deleter:   deleter,
		execCheck: execCheck,
		executor:  exec,
		timers:    make(map[uuid.UUID]*time.Timer),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start loads tasks from the DB, registers them, and starts scheduling.
func (s *Scheduler) Start() error {
	if err := s.loadCronTasks(); err != nil {
		return err
	}
	if err := s.loadScheduledTasks(); err != nil {
		return err
	}
	s.cron.Start()
	go s.heartbeatLoop()
	slog.Info("scheduler started")
	return nil
}

// Reload clears all entries and re-registers from the DB.
func (s *Scheduler) Reload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.Info("reloading scheduler")

	// Clear cron entries
	for _, entry := range s.cron.Entries() {
		s.cron.Remove(entry.ID)
	}

	// Clear pending timers
	for id, timer := range s.timers {
		timer.Stop()
		delete(s.timers, id)
	}

	if err := s.loadCronTasks(); err != nil {
		slog.Error("failed to reload cron tasks", "error", err)
	}
	if err := s.loadScheduledTasks(); err != nil {
		slog.Error("failed to reload scheduled tasks", "error", err)
	}
}

// Stop stops the scheduler and waits for all running jobs to finish.
func (s *Scheduler) Stop() {
	slog.Info("stopping scheduler")

	stopCtx := s.cron.Stop()
	s.cancel()

	// Stop all pending timers
	s.mu.Lock()
	for id, timer := range s.timers {
		timer.Stop()
		delete(s.timers, id)
	}
	s.mu.Unlock()

	<-stopCtx.Done()
	s.wg.Wait()

	slog.Info("scheduler stopped")
}

func (s *Scheduler) loadCronTasks() error {
	tasks, err := s.loader.ListEnabled(s.ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		task := task
		if task.CronExpression == nil {
			continue
		}
		_, err := s.cron.AddFunc(*task.CronExpression, func() {
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				s.executor.Run(s.ctx, task, nil)
			}()
		})
		if err != nil {
			slog.Error("failed to register cron job",
				"task_id", task.ID,
				"task_name", task.Name,
				"cron", *task.CronExpression,
				"error", err,
			)
			continue
		}
		slog.Info("registered cron job",
			"task_id", task.ID,
			"task_name", task.Name,
			"cron", *task.CronExpression,
		)
	}

	slog.Info("loaded cron tasks", "count", len(tasks))
	return nil
}

func (s *Scheduler) loadScheduledTasks() error {
	tasks, err := s.loader.ListPendingScheduled(s.ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		task := task
		if task.RunAt == nil {
			continue
		}

		delay := time.Until(*task.RunAt)
		if delay < 0 {
			// Past due — execute immediately
			delay = 0
		}

		timer := time.AfterFunc(delay, func() {
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()

				// Execute with auto-delete callback
				s.executor.Run(s.ctx, task, func() {
					if err := s.deleter.Delete(s.ctx, task.ID); err != nil {
						slog.Error("failed to auto-delete scheduled task",
							"task_id", task.ID,
							"error", err,
						)
					} else {
						slog.Info("auto-deleted scheduled task", "task_id", task.ID, "task_name", task.Name)
					}
				})

				// Clean up timer reference
				s.mu.Lock()
				delete(s.timers, task.ID)
				s.mu.Unlock()
			}()
		})

		s.timers[task.ID] = timer

		if delay == 0 {
			slog.Info("scheduled task past due, executing immediately",
				"task_id", task.ID,
				"task_name", task.Name,
				"run_at", task.RunAt,
			)
		} else {
			slog.Info("scheduled one-time task",
				"task_id", task.ID,
				"task_name", task.Name,
				"run_at", task.RunAt,
				"delay", delay.Round(time.Second),
			)
		}
	}

	if len(tasks) > 0 {
		slog.Info("loaded scheduled tasks", "count", len(tasks))
	}

	return nil
}

const (
	heartbeatInterval  = 10 * time.Minute
	staleExecutionAge  = 2 * time.Hour
	missedCronLookback = 24 * time.Hour
)

// heartbeatLoop runs every 10 minutes to recover missed cron jobs and mark stale executions.
func (s *Scheduler) heartbeatLoop() {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.runHeartbeat()
		}
	}
}

func (s *Scheduler) runHeartbeat() {
	slog.Info("heartbeat: checking for missed cron jobs and stale executions")

	// Mark stale running executions as failed
	if s.execCheck != nil {
		staleBefore := time.Now().Add(-staleExecutionAge)
		count, err := s.execCheck.MarkStaleExecutionsFailed(s.ctx, staleBefore)
		if err != nil {
			slog.Error("heartbeat: failed to mark stale executions", "error", err)
		} else if count > 0 {
			slog.Warn("heartbeat: marked stale executions as failed", "count", count)
		}
	}

	// Check for missed cron jobs
	tasks, err := s.loader.ListEnabled(s.ctx)
	if err != nil {
		slog.Error("heartbeat: failed to load enabled tasks", "error", err)
		return
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	now := time.Now()

	for _, task := range tasks {
		task := task
		if task.CronExpression == nil {
			continue
		}

		sched, err := parser.Parse(*task.CronExpression)
		if err != nil {
			continue
		}

		prevFire := prevFireTime(sched, now, missedCronLookback)
		if prevFire.IsZero() {
			continue
		}

		// Check if there's an execution at or after the previous expected fire time
		if s.execCheck != nil {
			latest, err := s.execCheck.GetLatestByAgentTaskID(s.ctx, task.ID)
			if err != nil {
				// No executions — task was missed
				slog.Warn("heartbeat: missed cron job detected (no executions), triggering",
					"task_id", task.ID,
					"task_name", task.Name,
					"expected_at", prevFire,
				)
				s.wg.Add(1)
				go func() {
					defer s.wg.Done()
					s.executor.Run(s.ctx, task, nil)
				}()
				continue
			}

			if latest.CreatedAt.Before(prevFire) {
				slog.Warn("heartbeat: missed cron job detected, triggering",
					"task_id", task.ID,
					"task_name", task.Name,
					"expected_at", prevFire,
					"last_execution_at", latest.CreatedAt,
				)
				s.wg.Add(1)
				go func() {
					defer s.wg.Done()
					s.executor.Run(s.ctx, task, nil)
				}()
			}
		}
	}
}

// prevFireTime computes the most recent time the cron should have fired before now.
func prevFireTime(sched cron.Schedule, now time.Time, lookback time.Duration) time.Time {
	t := now.Add(-lookback)
	var prev time.Time
	for {
		next := sched.Next(t)
		if next.After(now) {
			return prev
		}
		prev = next
		t = next
	}
}
