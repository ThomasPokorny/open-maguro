package executor

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"open-maguro/internal/domain"
	"open-maguro/internal/task_execution"
)

// ExecutionRepository defines the DB operations the executor needs.
type ExecutionRepository interface {
	Create(ctx context.Context, agentTaskID uuid.UUID, status domain.ExecutionStatus, taskName string) (*domain.TaskExecution, error)
	UpdateStatus(ctx context.Context, params task_execution.UpdateStatusParams) (*domain.TaskExecution, error)
}

type Executor struct {
	repo ExecutionRepository
}

func New(repo ExecutionRepository) *Executor {
	return &Executor{repo: repo}
}

// Run executes a single agent task. Safe to call from a goroutine.
// If onComplete is non-nil, it is called after execution finishes (used for auto-delete).
func (e *Executor) Run(ctx context.Context, task domain.AgentTask, onComplete func()) {
	logger := slog.With("task_id", task.ID, "task_name", task.Name)
	logger.Info("starting task execution")

	// Create execution record (pending)
	execution, err := e.repo.Create(ctx, task.ID, domain.StatusPending, task.Name)
	if err != nil {
		logger.Error("failed to create execution record", "error", err)
		return
	}

	execLogger := logger.With("execution_id", execution.ID)

	// Update to running
	now := time.Now()
	startedAt := pgtype.Timestamptz{Time: now, Valid: true}
	_, err = e.repo.UpdateStatus(ctx, task_execution.UpdateStatusParams{
		ID:        execution.ID,
		Status:    domain.StatusRunning,
		StartedAt: startedAt,
	})
	if err != nil {
		execLogger.Error("failed to update execution to running", "error", err)
		return
	}

	// Execute claude CLI with timeout
	timeout := time.Duration(task.TimeoutSeconds) * time.Second
	execCtx, execCancel := context.WithTimeout(ctx, timeout)
	defer execCancel()

	stdout, stderr, runErr := e.runClaude(execCtx, task.Prompt)

	// Record result (use parent ctx so DB write succeeds even after timeout)
	finishedAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	updateParams := task_execution.UpdateStatusParams{
		ID:         execution.ID,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
	}

	switch {
	case execCtx.Err() == context.DeadlineExceeded:
		updateParams.Status = domain.StatusTimeout
		updateParams.Error = pgtype.Text{
			String: fmt.Sprintf("execution timed out after %d seconds", task.TimeoutSeconds),
			Valid:  true,
		}
		if stdout != "" {
			updateParams.Summary = pgtype.Text{String: stdout, Valid: true}
		}
		execLogger.Warn("task execution timed out")

	case runErr != nil:
		updateParams.Status = domain.StatusFailure
		errMsg := stderr
		if errMsg == "" {
			errMsg = runErr.Error()
		}
		updateParams.Error = pgtype.Text{String: errMsg, Valid: true}
		if stdout != "" {
			updateParams.Summary = pgtype.Text{String: stdout, Valid: true}
		}
		execLogger.Error("task execution failed", "error", runErr)

	default:
		updateParams.Status = domain.StatusSuccess
		updateParams.Summary = pgtype.Text{String: stdout, Valid: true}
		execLogger.Info("task execution succeeded")
	}

	if _, err := e.repo.UpdateStatus(ctx, updateParams); err != nil {
		execLogger.Error("failed to update execution result", "error", err)
	}

	if onComplete != nil {
		onComplete()
	}
}

func (e *Executor) runClaude(ctx context.Context, prompt string) (stdout, stderr string, err error) {
	cmd := exec.CommandContext(ctx, "claude", "--print", "--output-format", "json", "-p", prompt)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}
