package kanban_executor

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
	"open-maguro/internal/executor"
	"open-maguro/internal/task_execution"
)

// KanbanRepository defines the DB operations the kanban executor needs.
type KanbanRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.KanbanTask, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.KanbanTaskStatus, result *string) (*domain.KanbanTask, error)
	ListPendingByAgentID(ctx context.Context, agentID uuid.UUID) ([]domain.KanbanTask, error)
	ListDistinctAgentsWithPending(ctx context.Context) ([]uuid.UUID, error)
	ResetInProgress(ctx context.Context) error
}

// AgentLoader loads agent task definitions.
type AgentLoader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AgentTask, error)
}

// ExecutionRepository creates and updates task execution records.
type ExecutionRepository interface {
	Create(ctx context.Context, agentTaskID uuid.UUID, status domain.ExecutionStatus, taskName string, triggeredByExecutionID *uuid.UUID) (*domain.TaskExecution, error)
	UpdateStatus(ctx context.Context, params task_execution.UpdateStatusParams) (*domain.TaskExecution, error)
}

type agentWorker struct {
	agentTaskID uuid.UUID
	queue       chan uuid.UUID
}

type KanbanExecutor struct {
	repo        KanbanRepository
	agentLoader AgentLoader
	execRepo    ExecutionRepository
	executor    *executor.Executor
	workers     map[uuid.UUID]*agentWorker
	mu          sync.Mutex
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func New(repo KanbanRepository, agentLoader AgentLoader, execRepo ExecutionRepository, exec *executor.Executor) *KanbanExecutor {
	ctx, cancel := context.WithCancel(context.Background())
	return &KanbanExecutor{
		repo:        repo,
		agentLoader: agentLoader,
		execRepo:    execRepo,
		executor:    exec,
		workers:     make(map[uuid.UUID]*agentWorker),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// LoadPending resets interrupted tasks and enqueues all pending kanban tasks.
// Called once at server startup.
func (ke *KanbanExecutor) LoadPending() error {
	// Reset any tasks stuck in "progress" from a previous crash
	if err := ke.repo.ResetInProgress(ke.ctx); err != nil {
		return fmt.Errorf("reset in-progress kanban tasks: %w", err)
	}

	agentIDs, err := ke.repo.ListDistinctAgentsWithPending(ke.ctx)
	if err != nil {
		return fmt.Errorf("list agents with pending kanban tasks: %w", err)
	}

	for _, agentID := range agentIDs {
		tasks, err := ke.repo.ListPendingByAgentID(ke.ctx, agentID)
		if err != nil {
			slog.Error("failed to load pending kanban tasks", "agent_task_id", agentID, "error", err)
			continue
		}

		for _, task := range tasks {
			ke.Enqueue(task)
		}

		slog.Info("loaded pending kanban tasks", "agent_task_id", agentID, "count", len(tasks))
	}

	return nil
}

// Enqueue dispatches a kanban task to the appropriate agent worker.
func (ke *KanbanExecutor) Enqueue(task domain.KanbanTask) {
	ke.mu.Lock()
	worker, exists := ke.workers[task.AgentTaskID]
	if !exists {
		worker = ke.startWorker(task.AgentTaskID)
		ke.workers[task.AgentTaskID] = worker
	}
	ke.mu.Unlock()

	select {
	case worker.queue <- task.ID:
	default:
		slog.Error("kanban worker queue full", "agent_task_id", task.AgentTaskID, "kanban_task_id", task.ID)
	}
}

// Stop cancels all workers and waits for them to finish their current task.
func (ke *KanbanExecutor) Stop() {
	slog.Info("stopping kanban executor")
	ke.cancel()
	ke.wg.Wait()
	slog.Info("kanban executor stopped")
}

func (ke *KanbanExecutor) startWorker(agentTaskID uuid.UUID) *agentWorker {
	w := &agentWorker{
		agentTaskID: agentTaskID,
		queue:       make(chan uuid.UUID, 1000),
	}

	ke.wg.Add(1)
	go ke.runWorker(w)

	return w
}

func (ke *KanbanExecutor) runWorker(w *agentWorker) {
	defer ke.wg.Done()
	logger := slog.With("agent_task_id", w.agentTaskID)
	logger.Info("kanban worker started")

	for {
		select {
		case <-ke.ctx.Done():
			logger.Info("kanban worker stopping")
			return
		case kanbanTaskID := <-w.queue:
			ke.processTask(w.agentTaskID, kanbanTaskID)
		}
	}
}

func (ke *KanbanExecutor) processTask(agentTaskID uuid.UUID, kanbanTaskID uuid.UUID) {
	logger := slog.With("agent_task_id", agentTaskID, "kanban_task_id", kanbanTaskID)

	// Fetch the kanban task
	kt, err := ke.repo.GetByID(ke.ctx, kanbanTaskID)
	if err != nil {
		logger.Error("failed to fetch kanban task", "error", err)
		return
	}

	// Set status to "progress"
	if _, err := ke.repo.UpdateStatus(ke.ctx, kanbanTaskID, domain.KanbanStatusProgress, nil); err != nil {
		logger.Error("failed to set kanban task to progress", "error", err)
		return
	}
	logger.Info("kanban task in progress", "title", kt.Title)

	// Load the agent task config
	agentTask, err := ke.agentLoader.GetByID(ke.ctx, agentTaskID)
	if err != nil {
		logger.Error("failed to load agent task", "error", err)
		result := "failed to load agent: " + err.Error()
		ke.repo.UpdateStatus(ke.ctx, kanbanTaskID, domain.KanbanStatusFailed, &result)
		return
	}

	// Create execution record (pending → running)
	taskName := fmt.Sprintf("[kanban] %s", kt.Title)
	execution, execErr := ke.execRepo.Create(ke.ctx, agentTaskID, domain.StatusPending, taskName, nil)
	if execErr != nil {
		logger.Error("failed to create execution record", "error", execErr)
	}

	now := time.Now()
	if execution != nil {
		if _, err := ke.execRepo.UpdateStatus(ke.ctx, task_execution.UpdateStatusParams{
			ID:        execution.ID,
			Status:    domain.StatusRunning,
			StartedAt: &now,
		}); err != nil {
			logger.Error("failed to update execution to running", "error", err)
		}
	}

	// Build prompt and run
	prompt := buildKanbanPrompt(agentTask, kt)
	stdout, stderr, runErr := ke.executor.RunKanban(ke.ctx, *agentTask, prompt)

	// Update kanban task status + execution record
	finishedAt := time.Now()
	if runErr != nil {
		result := stderr
		if result == "" {
			result = runErr.Error()
		}
		ke.repo.UpdateStatus(ke.ctx, kanbanTaskID, domain.KanbanStatusFailed, &result)
		logger.Error("kanban task failed", "title", kt.Title, "error", runErr)

		if execution != nil {
			summary := ptrIf(stdout, stdout != "")
			ke.execRepo.UpdateStatus(ke.ctx, task_execution.UpdateStatusParams{
				ID:         execution.ID,
				Status:     domain.StatusFailure,
				StartedAt:  &now,
				FinishedAt: &finishedAt,
				Error:      &result,
				Summary:    summary,
			})
		}
	} else {
		ke.repo.UpdateStatus(ke.ctx, kanbanTaskID, domain.KanbanStatusDone, &stdout)
		logger.Info("kanban task done", "title", kt.Title)

		if execution != nil {
			ke.execRepo.UpdateStatus(ke.ctx, task_execution.UpdateStatusParams{
				ID:         execution.ID,
				Status:     domain.StatusSuccess,
				StartedAt:  &now,
				FinishedAt: &finishedAt,
				Summary:    &stdout,
			})
		}
	}
}

func ptrIf(s string, cond bool) *string {
	if !cond {
		return nil
	}
	return &s
}

func buildKanbanPrompt(agent *domain.AgentTask, kt *domain.KanbanTask) string {
	var sb strings.Builder

	sb.WriteString(agent.Prompt)

	sb.WriteString("\n\n## Work Log\n")
	sb.WriteString("You have a `work-log.md` file in your workspace directory. ")
	sb.WriteString("Before starting work, read it to understand context from previous tasks. ")
	sb.WriteString("After completing this task, append a dated summary of what you did to `work-log.md`.\n")

	sb.WriteString("\n## Assigned Task\n")
	sb.WriteString("### ")
	sb.WriteString(kt.Title)
	sb.WriteString("\n")
	if kt.Description != "" {
		sb.WriteString(kt.Description)
		sb.WriteString("\n")
	}

	return sb.String()
}
