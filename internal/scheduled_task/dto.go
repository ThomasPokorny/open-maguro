package scheduled_task

import (
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type CreateRequest struct {
	Name         string    `json:"name"   validate:"required,min=1,max=255"`
	Prompt       string    `json:"prompt" validate:"required"`
	RunAt        time.Time `json:"run_at" validate:"required"`
	MCPConfig    *string   `json:"mcp_config"`
	AllowedTools *string   `json:"allowed_tools"`
}

type Response struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	TaskType          string     `json:"task_type"`
	Prompt            string     `json:"prompt"`
	RunAt             *time.Time `json:"run_at"`
	MCPConfig         *string    `json:"mcp_config,omitempty"`
	AllowedTools      *string    `json:"allowed_tools,omitempty"`
	Enabled           bool       `json:"enabled"`
	GlobalSkillAccess bool       `json:"global_skill_access"`
	CreatedAt         time.Time  `json:"created_at"`
}

func ToResponse(t *domain.AgentTask) Response {
	return Response{
		ID:                t.ID,
		Name:              t.Name,
		TaskType:          t.TaskType,
		Prompt:            t.Prompt,
		RunAt:             t.RunAt,
		MCPConfig:         t.MCPConfig,
		AllowedTools:      t.AllowedTools,
		Enabled:           t.Enabled,
		GlobalSkillAccess: t.GlobalSkillAccess,
		CreatedAt:         t.CreatedAt,
	}
}
