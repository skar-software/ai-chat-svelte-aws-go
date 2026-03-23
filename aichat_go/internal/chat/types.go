package chat

// StreamInput is the normalized input for the chat service (from API request).
type StreamInput struct {
	Input          string
	ConversationID string
	TenantID       string
	WorkspaceID    string
	UserID         string
	Model          string
}

// StreamEvent is a single normalized event sent to the client (SSE).
type StreamEvent struct {
	Event string
	Data  interface{}
}

// Usage holds token counts.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// MessageRecord is a single message for building provider context (history).
type MessageRecord struct {
	Role    string
	Content string
}
