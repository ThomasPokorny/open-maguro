package task_execution

import (
	"context"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TaskExecution, error)
	ListByAgentTaskID(ctx context.Context, agentTaskID uuid.UUID) ([]domain.TaskExecution, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.TaskExecution, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByAgentTaskID(ctx context.Context, agentTaskID uuid.UUID) ([]domain.TaskExecution, error) {
	return s.repo.ListByAgentTaskID(ctx, agentTaskID)
}
