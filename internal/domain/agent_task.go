package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	TaskTypeCron    = "cron"
	TaskTypeOneTime = "one_time"
)

type AgentTask struct {
	ID                uuid.UUID
	Name              string
	TaskType          string
	CronExpression    *string
	Prompt            string
	RunAt             *time.Time
	MCPConfig         *string
	AllowedTools      *string
	Enabled           bool
	SystemAgent       bool
	GlobalSkillAccess bool
	OnSuccessTaskID   *uuid.UUID
	OnFailureTaskID   *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
