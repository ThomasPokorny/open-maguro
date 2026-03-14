package kanban

import (
	"context"
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, params CreateRequest) (*domain.KanbanTask, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.KanbanTask, error)
	List(ctx context.Context) ([]domain.KanbanTask, error)
	ListByAgentID(ctx context.Context, agentID uuid.UUID) ([]domain.KanbanTask, error)
	ListByStatus(ctx context.Context, status domain.KanbanTaskStatus) ([]domain.KanbanTask, error)
	ListByAgentIDAndStatus(ctx context.Context, agentID uuid.UUID, status domain.KanbanTaskStatus) ([]domain.KanbanTask, error)
	ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]domain.KanbanTask, error)
	Update(ctx context.Context, id uuid.UUID, params UpdateRequest, existing *domain.KanbanTask) (*domain.KanbanTask, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*domain.KanbanTask, error) {
	return s.repo.Create(ctx, req)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.KanbanTask, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, agentID *uuid.UUID, status *string, teamID *uuid.UUID) ([]domain.KanbanTask, error) {
	var tasks []domain.KanbanTask
	var err error

	switch {
	case teamID != nil:
		tasks, err = s.repo.ListByTeamID(ctx, *teamID)
	case agentID != nil && status != nil:
		tasks, err = s.repo.ListByAgentIDAndStatus(ctx, *agentID, domain.KanbanTaskStatus(*status))
	case agentID != nil:
		tasks, err = s.repo.ListByAgentID(ctx, *agentID)
	case status != nil:
		tasks, err = s.repo.ListByStatus(ctx, domain.KanbanTaskStatus(*status))
	default:
		tasks, err = s.repo.List(ctx)
	}
	if err != nil {
		return nil, err
	}

	// Filter out done tasks older than 2 hours (unless explicitly filtering by status)
	if status == nil {
		cutoff := time.Now().Add(-2 * time.Hour)
		filtered := make([]domain.KanbanTask, 0, len(tasks))
		for _, t := range tasks {
			if t.Status == domain.KanbanStatusDone && t.UpdatedAt.Before(cutoff) {
				continue
			}
			filtered = append(filtered, t)
		}
		tasks = filtered
	}

	return tasks, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*domain.KanbanTask, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, req, existing)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
