package domain

import (
	"time"

	"github.com/google/uuid"
)

type KanbanTaskStatus string

const (
	KanbanStatusTodo     KanbanTaskStatus = "todo"
	KanbanStatusProgress KanbanTaskStatus = "progress"
	KanbanStatusDone     KanbanTaskStatus = "done"
	KanbanStatusFailed   KanbanTaskStatus = "failed"
)

type KanbanTask struct {
	ID          uuid.UUID
	Title       string
	Description string
	AgentTaskID uuid.UUID
	Status      KanbanTaskStatus
	Result      *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
