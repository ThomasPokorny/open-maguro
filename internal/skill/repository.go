package skill

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"open-maguro/internal/crypto"
	"open-maguro/internal/domain"
	"open-maguro/internal/sqlcgen"
)

type PostgresRepository struct {
	pool      *pgxpool.Pool
	queries   *sqlcgen.Queries
	secretKey []byte
}

func NewPostgresRepository(pool *pgxpool.Pool, secretKey []byte) *PostgresRepository {
	return &PostgresRepository{
		pool:      pool,
		queries:   sqlcgen.New(pool),
		secretKey: secretKey,
	}
}

func (r *PostgresRepository) Create(ctx context.Context, params CreateRequest) (*domain.Skill, error) {
	encSecrets, err := r.encryptSecrets(params.EnvironmentSecrets)
	if err != nil {
		return nil, fmt.Errorf("create skill: %w", err)
	}
	row, err := r.queries.CreateSkill(ctx, sqlcgen.CreateSkillParams{
		Title:              params.Title,
		Content:            params.Content,
		EnvironmentSecrets: encSecrets,
	})
	if err != nil {
		return nil, fmt.Errorf("create skill: %w", err)
	}
	return r.toDomain(row)
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Skill, error) {
	row, err := r.queries.GetSkill(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("skill not found: %s", id)
		}
		return nil, fmt.Errorf("get skill: %w", err)
	}
	return r.toDomain(row)
}

func (r *PostgresRepository) List(ctx context.Context) ([]domain.Skill, error) {
	rows, err := r.queries.ListSkills(ctx)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	skills := make([]domain.Skill, len(rows))
	for i, row := range rows {
		s, err := r.toDomain(row)
		if err != nil {
			return nil, err
		}
		skills[i] = *s
	}
	return skills, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, params UpdateRequest) (*domain.Skill, error) {
	var secrets map[string]string
	if params.EnvironmentSecrets != nil {
		secrets = *params.EnvironmentSecrets
	}
	encSecrets, err := r.encryptSecrets(secrets)
	if err != nil {
		return nil, fmt.Errorf("update skill: %w", err)
	}
	row, err := r.queries.UpdateSkill(ctx, sqlcgen.UpdateSkillParams{
		ID:                 id,
		Title:              *params.Title,
		Content:            *params.Content,
		EnvironmentSecrets: encSecrets,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("skill not found: %s", id)
		}
		return nil, fmt.Errorf("update skill: %w", err)
	}
	return r.toDomain(row)
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
		s, err := r.toDomain(row)
		if err != nil {
			return nil, err
		}
		skills[i] = *s
	}
	return skills, nil
}

func (r *PostgresRepository) encryptSecrets(secrets map[string]string) (pgtype.Text, error) {
	if len(secrets) == 0 {
		return pgtype.Text{}, nil
	}
	jsonBytes, err := json.Marshal(secrets)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("marshal secrets: %w", err)
	}
	encrypted, err := crypto.Encrypt(jsonBytes, r.secretKey)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("encrypt secrets: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(encrypted)
	return pgtype.Text{String: encoded, Valid: true}, nil
}

func (r *PostgresRepository) decryptSecrets(encrypted pgtype.Text) (map[string]string, error) {
	if !encrypted.Valid || encrypted.String == "" {
		return nil, nil
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted.String)
	if err != nil {
		return nil, fmt.Errorf("decode secrets: %w", err)
	}
	plaintext, err := crypto.Decrypt(ciphertext, r.secretKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt secrets: %w", err)
	}
	var secrets map[string]string
	if err := json.Unmarshal(plaintext, &secrets); err != nil {
		return nil, fmt.Errorf("unmarshal secrets: %w", err)
	}
	return secrets, nil
}

func (r *PostgresRepository) toDomain(row sqlcgen.Skill) (*domain.Skill, error) {
	secrets, err := r.decryptSecrets(row.EnvironmentSecrets)
	if err != nil {
		return nil, err
	}
	return &domain.Skill{
		ID:                 row.ID,
		Title:              row.Title,
		Content:            row.Content,
		EnvironmentSecrets: secrets,
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}, nil
}
