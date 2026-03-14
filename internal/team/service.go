package team

import (
	"context"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, params CreateRequest) (*domain.Team, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error)
	List(ctx context.Context) ([]domain.Team, error)
	Update(ctx context.Context, id uuid.UUID, params UpdateRequest) (*domain.Team, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*domain.Team, error) {
	return s.repo.Create(ctx, req)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]domain.Team, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*domain.Team, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
