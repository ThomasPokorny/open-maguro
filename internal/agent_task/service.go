package agent_task

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

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
	repo          Repository
	workspaceRoot string
}

func NewService(repo Repository, workspaceRoot string) *Service {
	return &Service{repo: repo, workspaceRoot: workspaceRoot}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*domain.AgentTask, error) {
	task, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	if s.workspaceRoot != "" {
		dir := filepath.Join(s.workspaceRoot, task.ID.String())
		if err := os.MkdirAll(dir, 0755); err != nil {
			slog.Error("failed to create agent workspace", "task_id", task.ID, "path", dir, "error", err)
		}
	}

	return task, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.AgentTask, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]domain.AgentTask, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*domain.AgentTask, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply partial updates
	merged := UpdateRequest{
		Name:              &existing.Name,
		CronExpression:    existing.CronExpression,
		Prompt:            &existing.Prompt,
		Enabled:           &existing.Enabled,
		MCPConfig:         existing.MCPConfig,
		AllowedTools:      existing.AllowedTools,
		SystemAgent:       &existing.SystemAgent,
		GlobalSkillAccess: &existing.GlobalSkillAccess,
		OnSuccessTaskID:   existing.OnSuccessTaskID,
		OnFailureTaskID:   existing.OnFailureTaskID,
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
	if req.MCPConfig != nil {
		merged.MCPConfig = req.MCPConfig
	}
	if req.AllowedTools != nil {
		merged.AllowedTools = req.AllowedTools
	}
	if req.SystemAgent != nil {
		merged.SystemAgent = req.SystemAgent
	}
	if req.GlobalSkillAccess != nil {
		merged.GlobalSkillAccess = req.GlobalSkillAccess
	}
	if req.OnSuccessTaskID != nil {
		merged.OnSuccessTaskID = req.OnSuccessTaskID
	}
	if req.OnFailureTaskID != nil {
		merged.OnFailureTaskID = req.OnFailureTaskID
	}

	// Validate no circular chains
	if err := s.validateNoChainCycle(ctx, id, merged.OnSuccessTaskID, merged.OnFailureTaskID); err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, id, merged)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if s.workspaceRoot != "" {
		dir := filepath.Join(s.workspaceRoot, id.String())
		if err := os.RemoveAll(dir); err != nil {
			slog.Error("failed to remove agent workspace", "task_id", id, "path", dir, "error", err)
		}
	}

	return nil
}

// validateNoChainCycle follows the chain from the given triggers and ensures
// it doesn't loop back to the source task id.
func (s *Service) validateNoChainCycle(ctx context.Context, sourceID uuid.UUID, onSuccess, onFailure *uuid.UUID) error {
	visited := map[uuid.UUID]bool{sourceID: true}

	var check func(id *uuid.UUID) error
	check = func(id *uuid.UUID) error {
		if id == nil {
			return nil
		}
		if visited[*id] {
			return fmt.Errorf("circular chain detected: task %s would create a cycle", id)
		}
		visited[*id] = true
		task, err := s.repo.GetByID(ctx, *id)
		if err != nil {
			return nil // target doesn't exist — not a cycle
		}
		if err := check(task.OnSuccessTaskID); err != nil {
			return err
		}
		return check(task.OnFailureTaskID)
	}

	if err := check(onSuccess); err != nil {
		return err
	}
	// Reset visited for failure path (only source is fixed)
	visited = map[uuid.UUID]bool{sourceID: true}
	return check(onFailure)
}
