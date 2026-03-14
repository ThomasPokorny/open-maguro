package team

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

func (r *PostgresRepository) Create(ctx context.Context, params CreateRequest) (*domain.Team, error) {
	color := params.Color
	if color == "" {
		color = "#6366f1"
	}

	row, err := r.queries.CreateTeam(ctx, sqlcgen.CreateTeamParams{
		ID:          uuid.New().String(),
		Title:       params.Title,
		Description: params.Description,
		Color:       color,
	})
	if err != nil {
		return nil, fmt.Errorf("create team: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error) {
	row, err := r.queries.GetTeam(ctx, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found: %s", id)
		}
		return nil, fmt.Errorf("get team: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]domain.Team, error) {
	rows, err := r.queries.ListTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}

	teams := make([]domain.Team, len(rows))
	for i, row := range rows {
		teams[i] = *toDomain(row)
	}
	return teams, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, params UpdateRequest) (*domain.Team, error) {
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	title := existing.Title
	if params.Title != nil {
		title = *params.Title
	}
	description := existing.Description
	if params.Description != nil {
		description = *params.Description
	}
	color := existing.Color
	if params.Color != nil {
		color = *params.Color
	}

	row, err := r.queries.UpdateTeam(ctx, sqlcgen.UpdateTeamParams{
		ID:          id.String(),
		Title:       title,
		Description: description,
		Color:       color,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found: %s", id)
		}
		return nil, fmt.Errorf("update team: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.DeleteTeam(ctx, id.String()); err != nil {
		return fmt.Errorf("delete team: %w", err)
	}
	return nil
}

func toDomain(row sqlcgen.Team) *domain.Team {
	return &domain.Team{
		ID:          uuid.MustParse(row.ID),
		Title:       row.Title,
		Description: row.Description,
		Color:       row.Color,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
