package domain

import (
	"time"

	"github.com/google/uuid"
)

type Skill struct {
	ID                 uuid.UUID
	Title              string
	Content            string
	EnvironmentSecrets map[string]string // decrypted; nil if no secrets
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
