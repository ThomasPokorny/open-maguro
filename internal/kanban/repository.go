package kanban

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
	"open-maguro/internal/sqlcgen"
)

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

func (r *PostgresRepository) Create(ctx context.Context, params CreateRequest) (*domain.KanbanTask, error) {
	row, err := r.queries.CreateKanbanTask(ctx, sqlcgen.CreateKanbanTaskParams{
		ID:          uuid.New().String(),
		Title:       params.Title,
		Description: params.Description,
		AgentTaskID: params.AgentTaskID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("create kanban task: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.KanbanTask, error) {
	row, err := r.queries.GetKanbanTask(ctx, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
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
	rows, err := r.queries.ListKanbanTasksByAgentID(ctx, agentID.String())
	if err != nil {
		return nil, fmt.Errorf("list kanban tasks by agent: %w", err)
	}
	return toDomainList(rows), nil
}

func (r *PostgresRepository) ListByStatus(ctx context.Context, status domain.KanbanTaskStatus) ([]domain.KanbanTask, error) {
	rows, err := r.queries.ListKanbanTasksByStatus(ctx, string(status))
	if err != nil {
		return nil, fmt.Errorf("list kanban tasks by status: %w", err)
	}
	return toDomainList(rows), nil
}

func (r *PostgresRepository) ListByAgentIDAndStatus(ctx context.Context, agentID uuid.UUID, status domain.KanbanTaskStatus) ([]domain.KanbanTask, error) {
	rows, err := r.queries.ListKanbanTasksByAgentIDAndStatus(ctx, sqlcgen.ListKanbanTasksByAgentIDAndStatusParams{
		AgentTaskID: agentID.String(),
		Status:      string(status),
	})
	if err != nil {
		return nil, fmt.Errorf("list kanban tasks by agent and status: %w", err)
	}
	return toDomainList(rows), nil
}

func (r *PostgresRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]domain.KanbanTask, error) {
	rows, err := r.queries.ListKanbanTasksByTeamID(ctx, sql.NullString{String: teamID.String(), Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list kanban tasks by team: %w", err)
	}
	return toDomainList(rows), nil
}

func (r *PostgresRepository) ListPendingByAgentID(ctx context.Context, agentID uuid.UUID) ([]domain.KanbanTask, error) {
	rows, err := r.queries.ListPendingKanbanTasksByAgentID(ctx, agentID.String())
	if err != nil {
		return nil, fmt.Errorf("list pending kanban tasks: %w", err)
	}
	return toDomainList(rows), nil
}

func (r *PostgresRepository) ListDistinctAgentsWithPending(ctx context.Context) ([]uuid.UUID, error) {
	ids, err := r.queries.ListDistinctAgentsWithPendingKanbanTasks(ctx)
	if err != nil {
		return nil, err
	}
	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		uuids[i] = uuid.MustParse(id)
	}
	return uuids, nil
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

	result := sql.NullString{}
	if existing.Result != nil {
		result = sql.NullString{String: *existing.Result, Valid: true}
	}

	row, err := r.queries.UpdateKanbanTask(ctx, sqlcgen.UpdateKanbanTaskParams{
		ID:          id.String(),
		Title:       title,
		Description: description,
		AgentTaskID: agentTaskID.String(),
		Status:      string(existing.Status),
		Result:      result,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("kanban task not found: %s", id)
		}
		return nil, fmt.Errorf("update kanban task: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.KanbanTaskStatus, result *string) (*domain.KanbanTask, error) {
	resultText := sql.NullString{}
	if result != nil {
		resultText = sql.NullString{String: *result, Valid: true}
	}
	row, err := r.queries.UpdateKanbanTaskStatus(ctx, sqlcgen.UpdateKanbanTaskStatusParams{
		ID:     id.String(),
		Status: string(status),
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
	return r.queries.DeleteKanbanTask(ctx, id.String())
}

func toDomain(row sqlcgen.KanbanTask) *domain.KanbanTask {
	t := &domain.KanbanTask{
		ID:          uuid.MustParse(row.ID),
		Title:       row.Title,
		Description: row.Description,
		AgentTaskID: uuid.MustParse(row.AgentTaskID),
		Status:      domain.KanbanTaskStatus(row.Status),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
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
