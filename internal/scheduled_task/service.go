package scheduled_task

import (
	"context"
	"time"

	"open-maguro/internal/domain"
)

type Repository interface {
	CreateScheduled(ctx context.Context, name, prompt string, runAt time.Time, timeoutSeconds int32) (*domain.AgentTask, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*domain.AgentTask, error) {
	timeout := int32(60)
	if req.TimeoutSeconds != nil {
		timeout = *req.TimeoutSeconds
	}
	return s.repo.CreateScheduled(ctx, req.Name, req.Prompt, req.RunAt, timeout)
}
