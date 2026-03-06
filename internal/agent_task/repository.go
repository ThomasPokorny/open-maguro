package agent_task

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"open-maguro/internal/domain"
	"open-maguro/internal/sqlcgen"
)

type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *sqlcgen.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool:    pool,
		queries: sqlcgen.New(pool),
	}
}

func (r *PostgresRepository) Create(ctx context.Context, params CreateRequest) (*domain.AgentTask, error) {
	enabled := true
	if params.Enabled != nil {
		enabled = *params.Enabled
	}
	timeout := int32(60)
	if params.TimeoutSeconds != nil {
		timeout = *params.TimeoutSeconds
	}

	mcpConfig := pgtype.Text{}
	if params.MCPConfig != nil {
		mcpConfig = pgtype.Text{String: *params.MCPConfig, Valid: true}
	}

	row, err := r.queries.CreateAgentTask(ctx, sqlcgen.CreateAgentTaskParams{
		Name:           params.Name,
		CronExpression: pgtype.Text{String: params.CronExpression, Valid: true},
		Prompt:         params.Prompt,
		Enabled:        enabled,
		TimeoutSeconds: timeout,
		McpConfig:      mcpConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("create agent task: %w", err)
	}

	return toDomain(row), nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AgentTask, error) {
	row, err := r.queries.GetAgentTask(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("agent task not found: %s", id)
		}
		return nil, fmt.Errorf("get agent task: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]domain.AgentTask, error) {
	rows, err := r.queries.ListAgentTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("list agent tasks: %w", err)
	}

	tasks := make([]domain.AgentTask, len(rows))
	for i, row := range rows {
		tasks[i] = *toDomain(row)
	}
	return tasks, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, params UpdateRequest) (*domain.AgentTask, error) {
	cronExpr := pgtype.Text{}
	if params.CronExpression != nil {
		cronExpr = pgtype.Text{String: *params.CronExpression, Valid: true}
	}
	mcpConfig := pgtype.Text{}
	if params.MCPConfig != nil {
		mcpConfig = pgtype.Text{String: *params.MCPConfig, Valid: true}
	}

	row, err := r.queries.UpdateAgentTask(ctx, sqlcgen.UpdateAgentTaskParams{
		ID:             id,
		Name:           *params.Name,
		CronExpression: cronExpr,
		Prompt:         *params.Prompt,
		Enabled:        *params.Enabled,
		TimeoutSeconds: *params.TimeoutSeconds,
		McpConfig:      mcpConfig,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("agent task not found: %s", id)
		}
		return nil, fmt.Errorf("update agent task: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteAgentTask(ctx, id)
	if err != nil {
		return fmt.Errorf("delete agent task: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListEnabled(ctx context.Context) ([]domain.AgentTask, error) {
	rows, err := r.queries.ListEnabledCronTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enabled cron tasks: %w", err)
	}

	tasks := make([]domain.AgentTask, len(rows))
	for i, row := range rows {
		tasks[i] = *toDomain(row)
	}
	return tasks, nil
}

func (r *PostgresRepository) ListPendingScheduled(ctx context.Context) ([]domain.AgentTask, error) {
	rows, err := r.queries.ListPendingScheduledTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pending scheduled tasks: %w", err)
	}

	tasks := make([]domain.AgentTask, len(rows))
	for i, row := range rows {
		tasks[i] = *toDomain(row)
	}
	return tasks, nil
}

func (r *PostgresRepository) CreateScheduled(ctx context.Context, name, prompt string, runAt time.Time, timeoutSeconds int32, mcpConfigVal *string) (*domain.AgentTask, error) {
	mcpConfig := pgtype.Text{}
	if mcpConfigVal != nil {
		mcpConfig = pgtype.Text{String: *mcpConfigVal, Valid: true}
	}

	row, err := r.queries.CreateScheduledTask(ctx, sqlcgen.CreateScheduledTaskParams{
		Name:           name,
		Prompt:         prompt,
		RunAt:          pgtype.Timestamptz{Time: runAt, Valid: true},
		TimeoutSeconds: timeoutSeconds,
		McpConfig:      mcpConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("create scheduled task: %w", err)
	}
	return toDomain(row), nil
}

func toDomain(row sqlcgen.AgentTask) *domain.AgentTask {
	task := &domain.AgentTask{
		ID:             row.ID,
		Name:           row.Name,
		TaskType:       row.TaskType,
		Prompt:         row.Prompt,
		Enabled:        row.Enabled,
		TimeoutSeconds: row.TimeoutSeconds,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
	if row.CronExpression.Valid {
		s := row.CronExpression.String
		task.CronExpression = &s
	}
	if row.RunAt.Valid {
		t := row.RunAt.Time
		task.RunAt = &t
	}
	if row.McpConfig.Valid {
		s := row.McpConfig.String
		task.MCPConfig = &s
	}
	return task
}
