package maguro_chat

type ChatRequest struct {
	Message string `json:"message" validate:"required,min=1"`
}

type ChatResponse struct {
	Reply     string `json:"reply"`
	SessionID string `json:"session_id,omitempty"`
}
