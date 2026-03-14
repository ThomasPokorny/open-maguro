package agent_task

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

func (r *PostgresRepository) Create(ctx context.Context, params CreateRequest) (*domain.AgentTask, error) {
	enabled := true
	if params.Enabled != nil {
		enabled = *params.Enabled
	}

	systemAgent := false
	if params.SystemAgent != nil {
		systemAgent = *params.SystemAgent
	}

	globalSkillAccess := false
	if params.GlobalSkillAccess != nil {
		globalSkillAccess = *params.GlobalSkillAccess
	}

	row, err := r.queries.CreateAgentTask(ctx, sqlcgen.CreateAgentTaskParams{
		ID:                uuid.New().String(),
		Name:              params.Name,
		CronExpression:    ptrToNullString(params.CronExpression),
		Prompt:            params.Prompt,
		Enabled:           enabled,
		McpConfig:         ptrToNullString(params.MCPConfig),
		AllowedTools:      ptrToNullString(params.AllowedTools),
		SystemAgent:       systemAgent,
		GlobalSkillAccess: globalSkillAccess,
		OnSuccessTaskID:   uuidPtrToNullString(params.OnSuccessTaskID),
		OnFailureTaskID:   uuidPtrToNullString(params.OnFailureTaskID),
		TeamID:            uuidPtrToNullString(params.TeamID),
	})
	if err != nil {
		return nil, fmt.Errorf("create agent task: %w", err)
	}

	return toDomain(row), nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AgentTask, error) {
	row, err := r.queries.GetAgentTask(ctx, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
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
		ID:                id.String(),
		Name:              *params.Name,
		CronExpression:    ptrToNullString(params.CronExpression),
		Prompt:            *params.Prompt,
		Enabled:           *params.Enabled,
		McpConfig:         ptrToNullString(params.MCPConfig),
		AllowedTools:      ptrToNullString(params.AllowedTools),
		SystemAgent:       boolPtrVal(params.SystemAgent),
		GlobalSkillAccess: boolPtrVal(params.GlobalSkillAccess),
		OnSuccessTaskID:   uuidPtrToNullString(params.OnSuccessTaskID),
		OnFailureTaskID:   uuidPtrToNullString(params.OnFailureTaskID),
		TeamID:            uuidPtrToNullString(params.TeamID),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent task not found: %s", id)
		}
		return nil, fmt.Errorf("update agent task: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteAgentTask(ctx, id.String())
	if err != nil {
		return fmt.Errorf("delete agent task: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]domain.AgentTask, error) {
	rows, err := r.queries.ListAgentTasksByTeamID(ctx, sql.NullString{String: teamID.String(), Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list agent tasks by team: %w", err)
	}

	tasks := make([]domain.AgentTask, len(rows))
	for i, row := range rows {
		tasks[i] = *toDomain(row)
	}
	return tasks, nil
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

func (r *PostgresRepository) CreateScheduled(ctx context.Context, name, prompt string, runAt time.Time, mcpConfigVal *string, allowedToolsVal *string) (*domain.AgentTask, error) {
	row, err := r.queries.CreateScheduledTask(ctx, sqlcgen.CreateScheduledTaskParams{
		ID:                uuid.New().String(),
		Name:              name,
		Prompt:            prompt,
		RunAt:             sql.NullTime{Time: runAt, Valid: true},
		McpConfig:         ptrToNullString(mcpConfigVal),
		AllowedTools:      ptrToNullString(allowedToolsVal),
		SystemAgent:       false,
		GlobalSkillAccess: false,
	})
	if err != nil {
		return nil, fmt.Errorf("create scheduled task: %w", err)
	}
	return toDomain(row), nil
}

func ptrToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func uuidPtrToNullString(id *uuid.UUID) sql.NullString {
	if id == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: id.String(), Valid: true}
}

func boolPtrVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func toDomain(row sqlcgen.AgentTask) *domain.AgentTask {
	task := &domain.AgentTask{
		ID:                uuid.MustParse(row.ID),
		Name:              row.Name,
		TaskType:          row.TaskType,
		Prompt:            row.Prompt,
		Enabled:           row.Enabled,
		SystemAgent:       row.SystemAgent,
		GlobalSkillAccess: row.GlobalSkillAccess,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
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
	if row.AllowedTools.Valid {
		s := row.AllowedTools.String
		task.AllowedTools = &s
	}
	if row.OnSuccessTaskID.Valid {
		id := uuid.MustParse(row.OnSuccessTaskID.String)
		task.OnSuccessTaskID = &id
	}
	if row.OnFailureTaskID.Valid {
		id := uuid.MustParse(row.OnFailureTaskID.String)
		task.OnFailureTaskID = &id
	}
	if row.TeamID.Valid {
		id := uuid.MustParse(row.TeamID.String)
		task.TeamID = &id
	}
	return task
}
