package skill

import (
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type CreateRequest struct {
	Title   string `json:"title"   validate:"required,min=1,max=255"`
	Content string `json:"content" validate:"required"`
}

type UpdateRequest struct {
	Title   *string `json:"title"   validate:"omitempty,min=1,max=255"`
	Content *string `json:"content"`
}

type Response struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToResponse(s *domain.Skill) Response {
	return Response{
		ID:        s.ID,
		Title:     s.Title,
		Content:   s.Content,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func ToResponseList(skills []domain.Skill) []Response {
	out := make([]Response, len(skills))
	for i := range skills {
		out[i] = ToResponse(&skills[i])
	}
	return out
}
