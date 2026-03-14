package domain

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID          uuid.UUID
	Title       string
	Description string
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
