package skill

import (
	"context"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, params CreateRequest) (*domain.Skill, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Skill, error)
	List(ctx context.Context) ([]domain.Skill, error)
	Update(ctx context.Context, id uuid.UUID, params UpdateRequest) (*domain.Skill, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AddAgentSkill(ctx context.Context, agentTaskID, skillID uuid.UUID) error
	RemoveAgentSkill(ctx context.Context, agentTaskID, skillID uuid.UUID) error
	ListByAgentTaskID(ctx context.Context, agentTaskID uuid.UUID) ([]domain.Skill, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*domain.Skill, error) {
	return s.repo.Create(ctx, req)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Skill, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]domain.Skill, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*domain.Skill, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existingSecrets := existing.EnvironmentSecrets
	merged := UpdateRequest{
		Title:              &existing.Title,
		Content:            &existing.Content,
		EnvironmentSecrets: &existingSecrets,
	}
	if req.Title != nil {
		merged.Title = req.Title
	}
	if req.Content != nil {
		merged.Content = req.Content
	}
	if req.EnvironmentSecrets != nil {
		merged.EnvironmentSecrets = req.EnvironmentSecrets
	}

	return s.repo.Update(ctx, id, merged)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) AddAgentSkill(ctx context.Context, agentTaskID, skillID uuid.UUID) error {
	return s.repo.AddAgentSkill(ctx, agentTaskID, skillID)
}

func (s *Service) RemoveAgentSkill(ctx context.Context, agentTaskID, skillID uuid.UUID) error {
	return s.repo.RemoveAgentSkill(ctx, agentTaskID, skillID)
}

func (s *Service) ListByAgentTaskID(ctx context.Context, agentTaskID uuid.UUID) ([]domain.Skill, error) {
	return s.repo.ListByAgentTaskID(ctx, agentTaskID)
}
