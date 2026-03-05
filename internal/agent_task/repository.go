package agent_task

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

	row, err := r.queries.CreateAgentTask(ctx, sqlcgen.CreateAgentTaskParams{
		Name:           params.Name,
		CronExpression: params.CronExpression,
		Prompt:         params.Prompt,
		Enabled:        enabled,
		TimeoutSeconds: timeout,
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
	row, err := r.queries.UpdateAgentTask(ctx, sqlcgen.UpdateAgentTaskParams{
		ID:             id,
		Name:           *params.Name,
		CronExpression: *params.CronExpression,
		Prompt:         *params.Prompt,
		Enabled:        *params.Enabled,
		TimeoutSeconds: *params.TimeoutSeconds,
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
	rows, err := r.queries.ListEnabledAgentTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enabled agent tasks: %w", err)
	}

	tasks := make([]domain.AgentTask, len(rows))
	for i, row := range rows {
		tasks[i] = *toDomain(row)
	}
	return tasks, nil
}

func toDomain(row sqlcgen.AgentTask) *domain.AgentTask {
	return &domain.AgentTask{
		ID:             row.ID,
		Name:           row.Name,
		CronExpression: row.CronExpression,
		Prompt:         row.Prompt,
		Enabled:        row.Enabled,
		TimeoutSeconds: row.TimeoutSeconds,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
}
