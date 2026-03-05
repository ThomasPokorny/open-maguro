package task_execution

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

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TaskExecution, error) {
	row, err := r.queries.GetTaskExecution(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("task execution not found: %s", id)
		}
		return nil, fmt.Errorf("get task execution: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) ListByAgentTaskID(ctx context.Context, agentTaskID uuid.UUID) ([]domain.TaskExecution, error) {
	rows, err := r.queries.ListTaskExecutionsByAgentTaskID(ctx, agentTaskID)
	if err != nil {
		return nil, fmt.Errorf("list task executions: %w", err)
	}

	executions := make([]domain.TaskExecution, len(rows))
	for i, row := range rows {
		executions[i] = *toDomain(row)
	}
	return executions, nil
}

func toDomain(row sqlcgen.TaskExecution) *domain.TaskExecution {
	exec := &domain.TaskExecution{
		ID:          row.ID,
		AgentTaskID: row.AgentTaskID,
		Status:      domain.ExecutionStatus(row.Status),
		CreatedAt:   row.CreatedAt.Time,
	}
	if row.StartedAt.Valid {
		t := row.StartedAt.Time
		exec.StartedAt = &t
	}
	if row.FinishedAt.Valid {
		t := row.FinishedAt.Time
		exec.FinishedAt = &t
	}
	if row.Summary.Valid {
		s := row.Summary.String
		exec.Summary = &s
	}
	if row.Error.Valid {
		s := row.Error.String
		exec.Error = &s
	}
	return exec
}
