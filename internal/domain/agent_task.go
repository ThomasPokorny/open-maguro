package domain

import (
	"time"

	"github.com/google/uuid"
)

type AgentTask struct {
	ID             uuid.UUID
	Name           string
	CronExpression string
	Prompt         string
	Enabled        bool
	TimeoutSeconds int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
