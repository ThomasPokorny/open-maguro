package task_execution

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
	"open-maguro/internal/sqlcgen"
)

// UpdateStatusParams holds the parameters for updating a task execution's status.
type UpdateStatusParams struct {
	ID         uuid.UUID
	Status     domain.ExecutionStatus
	StartedAt  *time.Time
	FinishedAt *time.Time
	Summary    *string
	Error      *string
}

type PostgresRepository struct {
	db      *sql.DB
	queries *sqlcgen.Queries
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db:      db,
		queries: sqlcgen.New(db),
	}
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TaskExecution, error) {
	row, err := r.queries.GetTaskExecution(ctx, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task execution not found: %s", id)
		}
		return nil, fmt.Errorf("get task execution: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]domain.TaskExecution, error) {
	rows, err := r.queries.ListTaskExecutions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list task executions: %w", err)
	}

	executions := make([]domain.TaskExecution, len(rows))
	for i, row := range rows {
		executions[i] = *toDomain(row)
	}
	return executions, nil
}

func (r *PostgresRepository) ListByAgentTaskID(ctx context.Context, agentTaskID uuid.UUID) ([]domain.TaskExecution, error) {
	rows, err := r.queries.ListTaskExecutionsByAgentTaskID(ctx, toNullString(agentTaskID.String()))
	if err != nil {
		return nil, fmt.Errorf("list task executions: %w", err)
	}

	executions := make([]domain.TaskExecution, len(rows))
	for i, row := range rows {
		executions[i] = *toDomain(row)
	}
	return executions, nil
}

func (r *PostgresRepository) Create(ctx context.Context, agentTaskID uuid.UUID, status domain.ExecutionStatus, taskName string, triggeredByExecutionID *uuid.UUID) (*domain.TaskExecution, error) {
	triggeredBy := sql.NullString{}
	if triggeredByExecutionID != nil {
		triggeredBy = sql.NullString{String: triggeredByExecutionID.String(), Valid: true}
	}
	id := uuid.New().String()
	row, err := r.queries.CreateTaskExecution(ctx, sqlcgen.CreateTaskExecutionParams{
		ID:                     id,
		AgentTaskID:            toNullString(agentTaskID.String()),
		Status:                 string(status),
		TaskName:               toNullString(taskName),
		TriggeredByExecutionID: triggeredBy,
	})
	if err != nil {
		return nil, fmt.Errorf("create task execution: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, params UpdateStatusParams) (*domain.TaskExecution, error) {
	row, err := r.queries.UpdateTaskExecutionStatus(ctx, sqlcgen.UpdateTaskExecutionStatusParams{
		ID:         params.ID.String(),
		Status:     string(params.Status),
		StartedAt:  toNullTime(params.StartedAt),
		FinishedAt: toNullTime(params.FinishedAt),
		Summary:    ptrToNullString(params.Summary),
		Error:      ptrToNullString(params.Error),
	})
	if err != nil {
		return nil, fmt.Errorf("update task execution status: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) GetLatestByAgentTaskID(ctx context.Context, agentTaskID uuid.UUID) (*domain.TaskExecution, error) {
	row, err := r.queries.GetLatestExecutionByAgentTaskID(ctx, toNullString(agentTaskID.String()))
	if err != nil {
		return nil, fmt.Errorf("get latest execution by agent task: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	count, err := r.queries.DeleteExecutionsOlderThan(ctx, before)
	if err != nil {
		return 0, fmt.Errorf("delete old executions: %w", err)
	}
	return count, nil
}

func (r *PostgresRepository) MarkStaleExecutionsFailed(ctx context.Context, staleBefore time.Time) (int, error) {
	rows, err := r.queries.ListStaleRunningExecutions(ctx, sql.NullTime{Time: staleBefore, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("list stale running executions: %w", err)
	}
	now := time.Now()
	errMsg := "marked as failed by heartbeat: execution appeared stale"
	for _, row := range rows {
		_, err := r.queries.UpdateTaskExecutionStatus(ctx, sqlcgen.UpdateTaskExecutionStatusParams{
			ID:         row.ID,
			Status:     string(domain.StatusFailure),
			StartedAt:  row.StartedAt,
			FinishedAt: sql.NullTime{Time: now, Valid: true},
			Error:      sql.NullString{String: errMsg, Valid: true},
		})
		if err != nil {
			return 0, fmt.Errorf("mark stale execution failed: %w", err)
		}
	}
	return len(rows), nil
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func ptrToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func toDomain(row sqlcgen.TaskExecution) *domain.TaskExecution {
	exec := &domain.TaskExecution{
		ID:        uuid.MustParse(row.ID),
		Status:    domain.ExecutionStatus(row.Status),
		CreatedAt: row.CreatedAt,
	}
	if row.AgentTaskID.Valid {
		id := uuid.MustParse(row.AgentTaskID.String)
		exec.AgentTaskID = &id
	}
	if row.TaskName.Valid {
		s := row.TaskName.String
		exec.TaskName = &s
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
	if row.TriggeredByExecutionID.Valid {
		id := uuid.MustParse(row.TriggeredByExecutionID.String)
		exec.TriggeredByExecutionID = &id
	}
	return exec
}
