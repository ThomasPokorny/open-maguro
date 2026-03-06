package agent_task

import (
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type CreateRequest struct {
	Name           string  `json:"name"            validate:"required,min=1,max=255"`
	CronExpression string  `json:"cron_expression" validate:"required"`
	Prompt         string  `json:"prompt"          validate:"required"`
	Enabled        *bool   `json:"enabled"`
	MCPConfig      *string `json:"mcp_config"`
}

type UpdateRequest struct {
	Name           *string `json:"name"            validate:"omitempty,min=1,max=255"`
	CronExpression *string `json:"cron_expression"`
	Prompt         *string `json:"prompt"`
	Enabled        *bool   `json:"enabled"`
	MCPConfig      *string `json:"mcp_config"`
}

type Response struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	TaskType       string     `json:"task_type"`
	CronExpression *string    `json:"cron_expression,omitempty"`
	Prompt         string     `json:"prompt"`
	RunAt          *time.Time `json:"run_at,omitempty"`
	MCPConfig      *string    `json:"mcp_config,omitempty"`
	Enabled        bool       `json:"enabled"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func ToResponse(t *domain.AgentTask) Response {
	return Response{
		ID:             t.ID,
		Name:           t.Name,
		TaskType:       t.TaskType,
		CronExpression: t.CronExpression,
		Prompt:         t.Prompt,
		RunAt:          t.RunAt,
		MCPConfig:      t.MCPConfig,
		Enabled:        t.Enabled,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}

func ToResponseList(tasks []domain.AgentTask) []Response {
	out := make([]Response, len(tasks))
	for i := range tasks {
		out[i] = ToResponse(&tasks[i])
	}
	return out
}
