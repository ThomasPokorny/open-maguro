package domain

import (
	"time"

	"github.com/google/uuid"
)

type ExecutionStatus string

const (
	StatusPending ExecutionStatus = "pending"
	StatusRunning ExecutionStatus = "running"
	StatusSuccess ExecutionStatus = "success"
	StatusFailure ExecutionStatus = "failure"
	StatusTimeout ExecutionStatus = "timeout"
)

type TaskExecution struct {
	ID          uuid.UUID
	AgentTaskID *uuid.UUID
	TaskName    *string
	Status      ExecutionStatus
	StartedAt   *time.Time
	FinishedAt  *time.Time
	Summary     *string
	Error       *string
	CreatedAt   time.Time
}
