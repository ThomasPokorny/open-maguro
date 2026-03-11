package skill

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

func (r *PostgresRepository) Create(ctx context.Context, params CreateRequest) (*domain.Skill, error) {
	row, err := r.queries.CreateSkill(ctx, sqlcgen.CreateSkillParams{
		Title:   params.Title,
		Content: params.Content,
	})
	if err != nil {
		return nil, fmt.Errorf("create skill: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Skill, error) {
	row, err := r.queries.GetSkill(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("skill not found: %s", id)
		}
		return nil, fmt.Errorf("get skill: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]domain.Skill, error) {
	rows, err := r.queries.ListSkills(ctx)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	skills := make([]domain.Skill, len(rows))
	for i, row := range rows {
		skills[i] = *toDomain(row)
	}
	return skills, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, params UpdateRequest) (*domain.Skill, error) {
	row, err := r.queries.UpdateSkill(ctx, sqlcgen.UpdateSkillParams{
		ID:      id,
		Title:   *params.Title,
		Content: *params.Content,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("skill not found: %s", id)
		}
		return nil, fmt.Errorf("update skill: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteSkill(ctx, id)
	if err != nil {
		return fmt.Errorf("delete skill: %w", err)
	}
	return nil
}

func (r *PostgresRepository) AddAgentSkill(ctx context.Context, agentTaskID, skillID uuid.UUID) error {
	return r.queries.AddAgentSkill(ctx, sqlcgen.AddAgentSkillParams{
		AgentTaskID: agentTaskID,
		SkillID:     skillID,
	})
}

func (r *PostgresRepository) RemoveAgentSkill(ctx context.Context, agentTaskID, skillID uuid.UUID) error {
	return r.queries.RemoveAgentSkill(ctx, sqlcgen.RemoveAgentSkillParams{
		AgentTaskID: agentTaskID,
		SkillID:     skillID,
	})
}

func (r *PostgresRepository) ListByAgentTaskID(ctx context.Context, agentTaskID uuid.UUID) ([]domain.Skill, error) {
	rows, err := r.queries.ListSkillsByAgentTaskID(ctx, agentTaskID)
	if err != nil {
		return nil, fmt.Errorf("list skills by agent task: %w", err)
	}
	skills := make([]domain.Skill, len(rows))
	for i, row := range rows {
		skills[i] = *toDomain(row)
	}
	return skills, nil
}

func toDomain(row sqlcgen.Skill) *domain.Skill {
	return &domain.Skill{
		ID:        row.ID,
		Title:     row.Title,
		Content:   row.Content,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
