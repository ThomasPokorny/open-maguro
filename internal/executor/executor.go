package executor

import (
	"bytes"
	"context"
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
	repo            ExecutionRepository
	globalMCPConfig string
	allowedTools    []string
}

func New(repo ExecutionRepository, globalMCPConfig string, allowedTools []string) *Executor {
	return &Executor{repo: repo, globalMCPConfig: globalMCPConfig, allowedTools: allowedTools}
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

	// Use task-level MCP config, fall back to global
	mcpConfig := e.globalMCPConfig
	if task.MCPConfig != nil {
		mcpConfig = *task.MCPConfig
	}

	// Execute claude CLI (no timeout — agents run until completion)
	stdout, stderr, runErr := e.runClaude(ctx, task.Prompt, mcpConfig)

	// Record result
	finishedAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	updateParams := task_execution.UpdateStatusParams{
		ID:         execution.ID,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
	}

	switch {
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

func (e *Executor) runClaude(ctx context.Context, prompt string, mcpConfig string) (stdout, stderr string, err error) {
	args := []string{"--print", "--output-format", "json"}
	if mcpConfig != "" {
		args = append(args, "--mcp-config", mcpConfig)
	}
	for _, tool := range e.allowedTools {
		args = append(args, "--allowedTools", tool)
	}
	args = append(args, "-p", prompt)

	cmd := exec.CommandContext(ctx, "claude", args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}
