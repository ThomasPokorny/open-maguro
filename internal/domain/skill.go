package domain

import (
	"time"

	"github.com/google/uuid"
)

type Skill struct {
	ID        uuid.UUID
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
