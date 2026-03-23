package api

import (
	"github.com/gofiber/fiber/v3"
	"aichat_go/internal/store"
)

func ListConversations(store *store.InMemoryStore) fiber.Handler {
	return func(c fiber.Ctx) error {
		tenantID := getString(c, "tenant_id", c.Query("tenant_id", "default"))
		list, err := store.ListConversations(tenantID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		out := make([]fiber.Map, 0, len(list))
		for _, conv := range list {
			out = append(out, fiber.Map{
				"id":           conv.ID,
				"tenant_id":    conv.TenantID,
				"workspace_id": conv.WorkspaceID,
				"user_id":      conv.UserID,
				"title":        conv.Title,
				"created_at":   conv.CreatedAt,
				"updated_at":   conv.UpdatedAt,
			})
		}
		return c.JSON(out)
	}
}

func GetConversationMessages(store *store.InMemoryStore) fiber.Handler {
	return func(c fiber.Ctx) error {
		convID := c.Params("id")
		if convID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "conversation id required"})
		}
		msgs, err := store.GetMessages(convID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		out := make([]fiber.Map, 0, len(msgs))
		for _, m := range msgs {
			out = append(out, fiber.Map{
				"id":              m.ID,
				"conversation_id": m.ConversationID,
				"role":            m.Role,
				"content":         m.Content,
				"provider":        m.Provider,
				"model":           m.Model,
				"input_tokens":    m.InputTokens,
				"output_tokens":   m.OutputTokens,
				"created_at":      m.CreatedAt,
			})
		}
		return c.JSON(out)
	}
}
