package task_execution

import (
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type Response struct {
	ID          uuid.UUID `json:"id"`
	AgentTaskID uuid.UUID `json:"agent_task_id"`
	Status      string    `json:"status"`
	StartedAt   *string   `json:"started_at,omitempty"`
	FinishedAt  *string   `json:"finished_at,omitempty"`
	Summary     *string   `json:"summary,omitempty"`
	Error       *string   `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func ToResponse(e *domain.TaskExecution) Response {
	resp := Response{
		ID:          e.ID,
		AgentTaskID: e.AgentTaskID,
		Status:      string(e.Status),
		Summary:     e.Summary,
		Error:       e.Error,
		CreatedAt:   e.CreatedAt,
	}
	if e.StartedAt != nil {
		s := e.StartedAt.Format(time.RFC3339)
		resp.StartedAt = &s
	}
	if e.FinishedAt != nil {
		s := e.FinishedAt.Format(time.RFC3339)
		resp.FinishedAt = &s
	}
	return resp
}

func ToResponseList(executions []domain.TaskExecution) []Response {
	out := make([]Response, len(executions))
	for i := range executions {
		out[i] = ToResponse(&executions[i])
	}
	return out
}
