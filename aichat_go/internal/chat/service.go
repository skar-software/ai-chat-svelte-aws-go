package chat

import (
	"context"
	"fmt"
	"log/slog"

	"aichat_go/internal/store"
)

type Service struct {
	store    *store.InMemoryStore
	provider OpenAIProvider
	logger   *slog.Logger
}

type OpenAIProvider interface {
	Stream(ctx context.Context, input *StreamInput, history []MessageRecord, out chan<- StreamEvent)
}

func NewChatService(s *store.InMemoryStore, p OpenAIProvider, logger *slog.Logger) *Service {
	return &Service{store: s, provider: p, logger: logger}
}

func (svc *Service) Stream(ctx context.Context, input *StreamInput, send func(StreamEvent)) error {
	convID := input.ConversationID
	tenantID := input.TenantID
	if tenantID == "" {
		tenantID = "default"
	}
	userID := input.UserID
	if userID == "" {
		userID = "anonymous"
	}
	workspaceID := input.WorkspaceID

	var conv *store.Conversation
	if convID != "" {
		var err error
		conv, err = svc.store.GetConversation(convID)
		if err != nil || conv == nil {
			svc.logger.Warn("conversation not found, creating new", "id", convID)
			convID = ""
		}
	}
	if convID == "" {
		var err error
		conv, err = svc.store.CreateConversation(tenantID, workspaceID, userID)
		if err != nil {
			send(StreamEvent{Event: "error", Data: map[string]string{"message": "failed to create conversation"}})
			return fmt.Errorf("create conversation: %w", err)
		}
		convID = conv.ID
	}

	if _, err := svc.store.AppendMessage(convID, "user", input.Input, "openai", input.Model, 0, 0); err != nil {
		send(StreamEvent{Event: "error", Data: map[string]string{"message": "failed to save message"}})
		return fmt.Errorf("append user message: %w", err)
	}

	msgs, err := svc.store.GetMessages(convID)
	if err != nil {
		svc.logger.Error("failed to load history", "error", err, "conversation_id", convID)
		send(StreamEvent{Event: "error", Data: map[string]string{"message": "failed to load history"}})
		return fmt.Errorf("get messages: %w", err)
	}

	history := make([]MessageRecord, 0, len(msgs))
	for _, m := range msgs {
		history = append(history, MessageRecord{Role: m.Role, Content: m.Content})
	}

	input.ConversationID = convID
	out := make(chan StreamEvent, 32)
	go svc.provider.Stream(ctx, input, history, out)

	var content string
	var inputTokens, outputTokens int
	parser := &StreamParser{state: "text"}

	for ev := range out {
		switch ev.Event {
		case "message.delta":
			if d, ok := ev.Data.(map[string]string); ok && d["delta"] != "" {
				forward, parts := parser.Feed(d["delta"])
				for _, s := range forward {
					content += s
					send(StreamEvent{Event: "message.delta", Data: map[string]string{"delta": s}})
				}
				for _, part := range parts {
					send(StreamEvent{Event: "part", Data: part})
				}
			}
		case "part":
			send(ev)
		case "usage":
			if u, ok := ev.Data.(map[string]int); ok {
				inputTokens = u["input_tokens"]
				outputTokens = u["output_tokens"]
				send(ev)
			}
		case "_stream_done":
			for _, s := range parser.Flush() {
				content += s
				send(StreamEvent{Event: "message.delta", Data: map[string]string{"delta": s}})
			}
			if _, err := svc.store.AppendMessage(convID, "assistant", content, "openai", input.Model, inputTokens, outputTokens); err != nil {
				svc.logger.Error("failed to persist assistant message", "error", err)
			}
			send(StreamEvent{Event: "message.completed", Data: map[string]string{"conversation_id": convID}})
		case "error":
			svc.logger.Error("provider stream error", "error", ev.Data)
			send(ev)
			return nil
		}
	}
	return nil
}
