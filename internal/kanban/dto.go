package kanban

import (
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type CreateRequest struct {
	Title       string    `json:"title"         validate:"required,min=1,max=255"`
	Description string    `json:"description"`
	AgentTaskID uuid.UUID `json:"agent_task_id" validate:"required"`
}

type UpdateRequest struct {
	Title       *string    `json:"title"         validate:"omitempty,min=1,max=255"`
	Description *string    `json:"description"`
	AgentTaskID *uuid.UUID `json:"agent_task_id"`
}

type Response struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	AgentTaskID uuid.UUID `json:"agent_task_id"`
	Status      string    `json:"status"`
	Result      *string   `json:"result,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToResponse(t *domain.KanbanTask) Response {
	return Response{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		AgentTaskID: t.AgentTaskID,
		Status:      string(t.Status),
		Result:      t.Result,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func ToResponseList(tasks []domain.KanbanTask) []Response {
	out := make([]Response, len(tasks))
	for i := range tasks {
		out[i] = ToResponse(&tasks[i])
	}
	return out
}
