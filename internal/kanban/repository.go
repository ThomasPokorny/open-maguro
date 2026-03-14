package kanban

import (
	"context"
	"fmt"

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

func (r *PostgresRepository) Create(ctx context.Context, params CreateRequest) (*domain.KanbanTask, error) {
	row, err := r.queries.CreateKanbanTask(ctx, sqlcgen.CreateKanbanTaskParams{
		Title:       params.Title,
		Description: params.Description,
		AgentTaskID: params.AgentTaskID,
	})
	if err != nil {
		return nil, fmt.Errorf("create kanban task: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.KanbanTask, error) {
	row, err := r.queries.GetKanbanTask(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("kanban task not found: %s", id)
		}
		return nil, fmt.Errorf("get kanban task: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]domain.KanbanTask, error) {
	rows, err := r.queries.ListKanbanTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("list kanban tasks: %w", err)
	}
	return toDomainList(rows), nil
}

func (r *PostgresRepository) ListByAgentID(ctx context.Context, agentID uuid.UUID) ([]domain.KanbanTask, error) {
	rows, err := r.queries.ListKanbanTasksByAgentID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("list kanban tasks by agent: %w", err)
	}
	return toDomainList(rows), nil
}

func (r *PostgresRepository) ListByStatus(ctx context.Context, status domain.KanbanTaskStatus) ([]domain.KanbanTask, error) {
	rows, err := r.queries.ListKanbanTasksByStatus(ctx, sqlcgen.KanbanTaskStatus(status))
	if err != nil {
		return nil, fmt.Errorf("list kanban tasks by status: %w", err)
	}
	return toDomainList(rows), nil
}

func (r *PostgresRepository) ListByAgentIDAndStatus(ctx context.Context, agentID uuid.UUID, status domain.KanbanTaskStatus) ([]domain.KanbanTask, error) {
	rows, err := r.queries.ListKanbanTasksByAgentIDAndStatus(ctx, sqlcgen.ListKanbanTasksByAgentIDAndStatusParams{
		AgentTaskID: agentID,
		Status:      sqlcgen.KanbanTaskStatus(status),
	})
	if err != nil {
		return nil, fmt.Errorf("list kanban tasks by agent and status: %w", err)
	}
	return toDomainList(rows), nil
}

func (r *PostgresRepository) ListPendingByAgentID(ctx context.Context, agentID uuid.UUID) ([]domain.KanbanTask, error) {
	rows, err := r.queries.ListPendingKanbanTasksByAgentID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("list pending kanban tasks: %w", err)
	}
	return toDomainList(rows), nil
}

func (r *PostgresRepository) ListDistinctAgentsWithPending(ctx context.Context) ([]uuid.UUID, error) {
	return r.queries.ListDistinctAgentsWithPendingKanbanTasks(ctx)
}

func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, params UpdateRequest, existing *domain.KanbanTask) (*domain.KanbanTask, error) {
	title := existing.Title
	if params.Title != nil {
		title = *params.Title
	}
	description := existing.Description
	if params.Description != nil {
		description = *params.Description
	}
	agentTaskID := existing.AgentTaskID
	if params.AgentTaskID != nil {
		agentTaskID = *params.AgentTaskID
	}

	result := pgtype.Text{}
	if existing.Result != nil {
		result = pgtype.Text{String: *existing.Result, Valid: true}
	}

	row, err := r.queries.UpdateKanbanTask(ctx, sqlcgen.UpdateKanbanTaskParams{
		ID:          id,
		Title:       title,
		Description: description,
		AgentTaskID: agentTaskID,
		Status:      sqlcgen.KanbanTaskStatus(existing.Status),
		Result:      result,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("kanban task not found: %s", id)
		}
		return nil, fmt.Errorf("update kanban task: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.KanbanTaskStatus, result *string) (*domain.KanbanTask, error) {
	resultText := pgtype.Text{}
	if result != nil {
		resultText = pgtype.Text{String: *result, Valid: true}
	}
	row, err := r.queries.UpdateKanbanTaskStatus(ctx, sqlcgen.UpdateKanbanTaskStatusParams{
		ID:     id,
		Status: sqlcgen.KanbanTaskStatus(status),
		Result: resultText,
	})
	if err != nil {
		return nil, fmt.Errorf("update kanban task status: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) ResetInProgress(ctx context.Context) error {
	return r.queries.ResetInProgressKanbanTasks(ctx)
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteKanbanTask(ctx, id)
}

func toDomain(row sqlcgen.KanbanTask) *domain.KanbanTask {
	t := &domain.KanbanTask{
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		AgentTaskID: row.AgentTaskID,
		Status:      domain.KanbanTaskStatus(row.Status),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
	if row.Result.Valid {
		s := row.Result.String
		t.Result = &s
	}
	return t
}

func toDomainList(rows []sqlcgen.KanbanTask) []domain.KanbanTask {
	tasks := make([]domain.KanbanTask, len(rows))
	for i, row := range rows {
		tasks[i] = *toDomain(row)
	}
	return tasks
}
