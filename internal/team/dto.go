package team

import (
	"time"

	"github.com/google/uuid"
	"open-maguro/internal/domain"
)

type CreateRequest struct {
	Title       string `json:"title"       validate:"required,min=1,max=255"`
	Description string `json:"description"`
	Color       string `json:"color"       validate:"omitempty,hexcolor"`
}

type UpdateRequest struct {
	Title       *string `json:"title"       validate:"omitempty,min=1,max=255"`
	Description *string `json:"description"`
	Color       *string `json:"color"       validate:"omitempty,hexcolor"`
}

type Response struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToResponse(t *domain.Team) Response {
	return Response{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Color:       t.Color,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func ToResponseList(teams []domain.Team) []Response {
	out := make([]Response, len(teams))
	for i := range teams {
		out[i] = ToResponse(&teams[i])
	}
	return out
}
