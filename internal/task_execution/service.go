package task_execution

import (
	"context"
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TaskExecution, error)
	List(ctx context.Context) ([]domain.TaskExecution, error)
	ListByAgentTaskID(ctx context.Context, agentTaskID uuid.UUID) ([]domain.TaskExecution, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
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

func (s *Service) List(ctx context.Context) ([]domain.TaskExecution, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListByAgentTaskID(ctx context.Context, agentTaskID uuid.UUID) ([]domain.TaskExecution, error) {
	return s.repo.ListByAgentTaskID(ctx, agentTaskID)
}

func (s *Service) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	return s.repo.DeleteOlderThan(ctx, before)
}
