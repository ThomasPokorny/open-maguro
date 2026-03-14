package skill

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type CreateRequest struct {
	Title              string            `json:"title"               validate:"required,min=1,max=255"`
	Content            string            `json:"content"             validate:"required"`
	EnvironmentSecrets map[string]string `json:"environment_secrets"`
}

type UpdateRequest struct {
	Title              *string            `json:"title"               validate:"omitempty,min=1,max=255"`
	Content            *string            `json:"content"`
	EnvironmentSecrets *map[string]string `json:"environment_secrets"`
}

type Response struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	SecretKeys []string  `json:"secret_keys"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func ToResponse(s *domain.Skill) Response {
	keys := make([]string, 0, len(s.EnvironmentSecrets))
	for k := range s.EnvironmentSecrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return Response{
		ID:         s.ID,
		Title:      s.Title,
		Content:    s.Content,
		SecretKeys: keys,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

func ToResponseList(skills []domain.Skill) []Response {
	out := make([]Response, len(skills))
	for i := range skills {
		out[i] = ToResponse(&skills[i])
	}
	return out
}
