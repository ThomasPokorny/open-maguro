package executor

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"open-maguro/internal/domain"
	"open-maguro/internal/task_execution"
)

// ExecutionRepository defines the DB operations the executor needs.
type ExecutionRepository interface {
	Create(ctx context.Context, agentTaskID uuid.UUID, status domain.ExecutionStatus, taskName string, triggeredByExecutionID *uuid.UUID) (*domain.TaskExecution, error)
	UpdateStatus(ctx context.Context, params task_execution.UpdateStatusParams) (*domain.TaskExecution, error)
}

// SkillLoader loads skills needed for a task execution.
type SkillLoader interface {
	ListByAgentTaskID(ctx context.Context, agentTaskID uuid.UUID) ([]domain.Skill, error)
	List(ctx context.Context) ([]domain.Skill, error)
}

// TaskLoader loads agent tasks (used for chaining).
type TaskLoader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AgentTask, error)
}

const systemPrompt = `# OpenMaguro🐟 Agent Orchestrator.
You are an agent orchestrated by the OpenMaguro🐟 Agent Orchestrator project. Similarly to OpenClaw, users can create agents and schedule them fulfilling different tasks. This is a task running in the background. So there is no means of getting additional tool calls whitelisted. Try to fulfill the user request by all means.`

type Executor struct {
	repo            ExecutionRepository
	skillLoader     SkillLoader
	taskLoader      TaskLoader
	globalMCPConfig string
	allowedTools    []string
	workspaceRoot   string
}

func New(repo ExecutionRepository, skillLoader SkillLoader, taskLoader TaskLoader, globalMCPConfig string, allowedTools []string, workspaceRoot string) *Executor {
	return &Executor{repo: repo, skillLoader: skillLoader, taskLoader: taskLoader, globalMCPConfig: globalMCPConfig, allowedTools: allowedTools, workspaceRoot: workspaceRoot}
}

// Run executes a single agent task. Safe to call from a goroutine.
// If onComplete is non-nil, it is called after execution finishes (used for auto-delete).
func (e *Executor) Run(ctx context.Context, task domain.AgentTask, onComplete func()) {
	e.runInternal(ctx, task, nil, "", onComplete)
}

// runInternal handles execution with optional chain context.
func (e *Executor) runInternal(ctx context.Context, task domain.AgentTask, triggeredByExecID *uuid.UUID, chainContext string, onComplete func()) {
	logger := slog.With("task_id", task.ID, "task_name", task.Name)
	logger.Info("starting task execution")

	// Create execution record (pending)
	execution, err := e.repo.Create(ctx, task.ID, domain.StatusPending, task.Name, triggeredByExecID)
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

	// Use task-level MCP config, fall back to global.
	// Resolve to absolute path so it works regardless of cmd.Dir (workspace).
	mcpConfig := e.globalMCPConfig
	if task.MCPConfig != nil {
		mcpConfig = *task.MCPConfig
	}
	if mcpConfig != "" && !filepath.IsAbs(mcpConfig) {
		if abs, err := filepath.Abs(mcpConfig); err == nil {
			mcpConfig = abs
		}
	}

	// Merge global + per-task allowed tools (additive)
	var taskTools []string
	if task.AllowedTools != nil && *task.AllowedTools != "" {
		taskTools = strings.Split(*task.AllowedTools, ",")
	}

	// Load and inject skills into prompt
	prompt := task.Prompt
	if e.skillLoader != nil {
		skills, err := e.loadSkills(ctx, task)
		if err != nil {
			execLogger.Error("failed to load skills", "error", err)
		} else if len(skills) > 0 {
			prompt = e.buildPromptWithSkills(skills, task.Prompt)
		}
	}

	// Prepend chain context if this was triggered by another agent
	if chainContext != "" {
		prompt = chainContext + "\n\n" + prompt
	}

	// Prepare workspace directory (ensure it exists for pre-existing agents)
	workspaceDir := ""
	if e.workspaceRoot != "" {
		workspaceDir = filepath.Join(e.workspaceRoot, task.ID.String())
		if err := os.MkdirAll(workspaceDir, 0755); err != nil {
			execLogger.Error("failed to ensure workspace directory", "path", workspaceDir, "error", err)
			workspaceDir = "" // fall back to no workspace
		}
	}

	// Prepend system prompt (with workspace info if available)
	sp := systemPrompt
	if workspaceDir != "" {
		sp += fmt.Sprintf("\n\nYour workspace directory is: %s\nYou can read and write files here freely. Files persist between runs.", workspaceDir)
		// Inject file tools so the agent can actually read/write in its workspace
		taskTools = append(taskTools, "Read", "Write", "Edit", "Glob", "Grep")
	}
	prompt = sp + "\n\n" + prompt

	// Execute claude CLI (no timeout — agents run until completion)
	stdout, stderr, runErr := e.runClaude(ctx, prompt, mcpConfig, taskTools, workspaceDir)

	// Record result
	finishedAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	updateParams := task_execution.UpdateStatusParams{
		ID:         execution.ID,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
	}

	var finalStatus domain.ExecutionStatus
	switch {
	case runErr != nil:
		finalStatus = domain.StatusFailure
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
		finalStatus = domain.StatusSuccess
		updateParams.Status = domain.StatusSuccess
		updateParams.Summary = pgtype.Text{String: stdout, Valid: true}
		execLogger.Info("task execution succeeded")
	}

	if _, err := e.repo.UpdateStatus(ctx, updateParams); err != nil {
		execLogger.Error("failed to update execution result", "error", err)
	}

	// Trigger chained agent if configured
	e.triggerChain(ctx, task, execution.ID, finalStatus, stdout, stderr)

	if onComplete != nil {
		onComplete()
	}
}

// triggerChain fires the on_success or on_failure chained agent if configured.
func (e *Executor) triggerChain(ctx context.Context, task domain.AgentTask, executionID uuid.UUID, status domain.ExecutionStatus, stdout, stderr string) {
	if e.taskLoader == nil {
		return
	}

	var nextTaskID *uuid.UUID
	switch status {
	case domain.StatusSuccess:
		nextTaskID = task.OnSuccessTaskID
	case domain.StatusFailure:
		nextTaskID = task.OnFailureTaskID
	}

	if nextTaskID == nil {
		return
	}

	nextTask, err := e.taskLoader.GetByID(ctx, *nextTaskID)
	if err != nil {
		slog.Error("failed to load chained task",
			"task_id", task.ID,
			"chained_task_id", nextTaskID,
			"error", err,
		)
		return
	}

	// Build chain context with parent output
	var chainCtx string
	output := stdout
	if status == domain.StatusFailure {
		output = stderr
		if output == "" {
			output = "(no error output)"
		}
	}
	chainCtx = fmt.Sprintf("This task was triggered by agent %q (execution: %s, status: %s).\n\nOutput from the triggering agent:\n---\n%s\n---",
		task.Name, executionID, status, output)

	slog.Info("triggering chained agent",
		"parent_task", task.Name,
		"chained_task", nextTask.Name,
		"trigger", string(status),
	)

	go e.runInternal(ctx, *nextTask, &executionID, chainCtx, nil)
}

func (e *Executor) runClaude(ctx context.Context, prompt string, mcpConfig string, extraTools []string, workingDir string) (stdout, stderr string, err error) {
	args := []string{"--print", "--output-format", "json"}
	if mcpConfig != "" {
		args = append(args, "--mcp-config", mcpConfig)
	}
	for _, tool := range e.allowedTools {
		args = append(args, "--allowedTools", tool)
	}
	for _, tool := range extraTools {
		tool = strings.TrimSpace(tool)
		if tool != "" {
			args = append(args, "--allowedTools", tool)
		}
	}
	args = append(args, "-p", prompt)

	//
	cmd := exec.CommandContext(ctx, "claude", args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}

// RunKanban executes an agent with a custom kanban prompt. It handles MCP config,
// allowed tools, skills, workspace, and system prompt — but does NOT create a
// task_execution record or trigger chaining. Returns stdout, stderr, and error.
func (e *Executor) RunKanban(ctx context.Context, task domain.AgentTask, kanbanPrompt string) (stdout, stderr string, err error) {
	// MCP config
	mcpConfig := e.globalMCPConfig
	if task.MCPConfig != nil {
		mcpConfig = *task.MCPConfig
	}
	if mcpConfig != "" && !filepath.IsAbs(mcpConfig) {
		if abs, err := filepath.Abs(mcpConfig); err == nil {
			mcpConfig = abs
		}
	}

	// Merge tools
	var taskTools []string
	if task.AllowedTools != nil && *task.AllowedTools != "" {
		taskTools = strings.Split(*task.AllowedTools, ",")
	}

	// Load skills
	prompt := kanbanPrompt
	if e.skillLoader != nil {
		skills, err := e.loadSkills(ctx, task)
		if err == nil && len(skills) > 0 {
			prompt = e.buildPromptWithSkills(skills, kanbanPrompt)
		}
	}

	// Workspace
	workspaceDir := ""
	if e.workspaceRoot != "" {
		workspaceDir = filepath.Join(e.workspaceRoot, task.ID.String())
		if mkErr := os.MkdirAll(workspaceDir, 0755); mkErr != nil {
			workspaceDir = ""
		}
	}

	// System prompt
	sp := systemPrompt
	if workspaceDir != "" {
		sp += fmt.Sprintf("\n\nYour workspace directory is: %s\nYou can read and write files here freely. Files persist between runs.", workspaceDir)
		taskTools = append(taskTools, "Read", "Write", "Edit", "Glob", "Grep")
	}
	prompt = sp + "\n\n" + prompt

	return e.runClaude(ctx, prompt, mcpConfig, taskTools, workspaceDir)
}

func (e *Executor) loadSkills(ctx context.Context, task domain.AgentTask) ([]domain.Skill, error) {
	if task.GlobalSkillAccess {
		return e.skillLoader.List(ctx)
	}
	return e.skillLoader.ListByAgentTaskID(ctx, task.ID)
}

func (e *Executor) buildPromptWithSkills(skills []domain.Skill, taskPrompt string) string {
	var sb strings.Builder
	sb.WriteString("These are your skills:\n\n")
	for i, s := range skills {
		sb.WriteString("## ")
		sb.WriteString(s.Title)
		sb.WriteString("\n\n")
		sb.WriteString(s.Content)
		if i < len(skills)-1 {
			sb.WriteString("\n\n---\n\n")
		}
	}
	sb.WriteString("\n\n---\n\nYour task:\n")
	sb.WriteString(taskPrompt)
	return sb.String()
}
