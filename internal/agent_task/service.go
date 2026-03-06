package agent_task

import (
	"context"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, params CreateRequest) (*domain.AgentTask, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AgentTask, error)
	List(ctx context.Context) ([]domain.AgentTask, error)
	Update(ctx context.Context, id uuid.UUID, params UpdateRequest) (*domain.AgentTask, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*domain.AgentTask, error) {
	return s.repo.Create(ctx, req)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.AgentTask, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]domain.AgentTask, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*domain.AgentTask, error) {
	// For update, we first get the existing task, apply changes, then save
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply partial updates
	merged := UpdateRequest{
		Name:           &existing.Name,
		CronExpression: existing.CronExpression,
		Prompt:         &existing.Prompt,
		Enabled:        &existing.Enabled,
		TimeoutSeconds: &existing.TimeoutSeconds,
		MCPConfig:      existing.MCPConfig,
	}
	if req.Name != nil {
		merged.Name = req.Name
	}
	if req.CronExpression != nil {
		merged.CronExpression = req.CronExpression
	}
	if req.Prompt != nil {
		merged.Prompt = req.Prompt
	}
	if req.Enabled != nil {
		merged.Enabled = req.Enabled
	}
	if req.TimeoutSeconds != nil {
		merged.TimeoutSeconds = req.TimeoutSeconds
	}
	if req.MCPConfig != nil {
		merged.MCPConfig = req.MCPConfig
	}

	return s.repo.Update(ctx, id, merged)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
