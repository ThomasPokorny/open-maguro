package scheduled_task

import (
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type CreateRequest struct {
	Name           string    `json:"name"            validate:"required,min=1,max=255"`
	Prompt         string    `json:"prompt"          validate:"required"`
	RunAt          time.Time `json:"run_at"          validate:"required"`
	TimeoutSeconds *int32    `json:"timeout_seconds" validate:"omitempty,min=1,max=3600"`
}

type Response struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	TaskType       string     `json:"task_type"`
	Prompt         string     `json:"prompt"`
	RunAt          *time.Time `json:"run_at"`
	Enabled        bool       `json:"enabled"`
	TimeoutSeconds int32      `json:"timeout_seconds"`
	CreatedAt      time.Time  `json:"created_at"`
}

func ToResponse(t *domain.AgentTask) Response {
	return Response{
		ID:             t.ID,
		Name:           t.Name,
		TaskType:       t.TaskType,
		Prompt:         t.Prompt,
		RunAt:          t.RunAt,
		Enabled:        t.Enabled,
		TimeoutSeconds: t.TimeoutSeconds,
		CreatedAt:      t.CreatedAt,
	}
}
