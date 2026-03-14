package agent_task

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type CreateRequest struct {
	Name              string     `json:"name"            validate:"required,min=1,max=255"`
	CronExpression    *string    `json:"cron_expression"`
	Prompt            string     `json:"prompt"          validate:"required"`
	Enabled           *bool      `json:"enabled"`
	MCPConfig         *string    `json:"mcp_config"`
	AllowedTools      *string    `json:"allowed_tools"`
	SystemAgent       *bool      `json:"system_agent"`
	GlobalSkillAccess *bool      `json:"global_skill_access"`
	OnSuccessTaskID   *uuid.UUID `json:"on_success_task_id"`
	OnFailureTaskID   *uuid.UUID `json:"on_failure_task_id"`
	TeamID            *uuid.UUID `json:"team_id"`
}

type UpdateRequest struct {
	Name              *string    `json:"name"            validate:"omitempty,min=1,max=255"`
	CronExpression    *string    `json:"cron_expression"`
	Prompt            *string    `json:"prompt"`
	Enabled           *bool      `json:"enabled"`
	MCPConfig         *string    `json:"mcp_config"`
	AllowedTools      *string    `json:"allowed_tools"`
	SystemAgent       *bool      `json:"system_agent"`
	GlobalSkillAccess *bool      `json:"global_skill_access"`
	OnSuccessTaskID   *uuid.UUID `json:"on_success_task_id"`
	OnFailureTaskID   *uuid.UUID `json:"on_failure_task_id"`
	TeamID            *uuid.UUID `json:"team_id"`
	TeamIDSet         bool       `json:"-"` // true when team_id key was present in JSON (allows explicit null)
}

func (u *UpdateRequest) UnmarshalJSON(data []byte) error {
	type Alias UpdateRequest
	aux := &struct{ *Alias }{Alias: (*Alias)(u)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if _, ok := raw["team_id"]; ok {
		u.TeamIDSet = true
	}
	return nil
}

type Response struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	TaskType          string     `json:"task_type"`
	CronExpression    *string    `json:"cron_expression,omitempty"`
	Prompt            string     `json:"prompt"`
	RunAt             *time.Time `json:"run_at,omitempty"`
	MCPConfig         *string    `json:"mcp_config,omitempty"`
	AllowedTools      *string    `json:"allowed_tools,omitempty"`
	Enabled           bool       `json:"enabled"`
	SystemAgent       bool       `json:"system_agent"`
	GlobalSkillAccess bool       `json:"global_skill_access"`
	OnSuccessTaskID   *uuid.UUID `json:"on_success_task_id,omitempty"`
	OnFailureTaskID   *uuid.UUID `json:"on_failure_task_id,omitempty"`
	TeamID            *uuid.UUID `json:"team_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func ToResponse(t *domain.AgentTask) Response {
	return Response{
		ID:                t.ID,
		Name:              t.Name,
		TaskType:          t.TaskType,
		CronExpression:    t.CronExpression,
		Prompt:            t.Prompt,
		RunAt:             t.RunAt,
		MCPConfig:         t.MCPConfig,
		AllowedTools:      t.AllowedTools,
		Enabled:           t.Enabled,
		SystemAgent:       t.SystemAgent,
		GlobalSkillAccess: t.GlobalSkillAccess,
		OnSuccessTaskID:   t.OnSuccessTaskID,
		OnFailureTaskID:   t.OnFailureTaskID,
		TeamID:            t.TeamID,
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
	}
}

func ToResponseList(tasks []domain.AgentTask) []Response {
	out := make([]Response, len(tasks))
	for i := range tasks {
		out[i] = ToResponse(&tasks[i])
	}
	return out
}
