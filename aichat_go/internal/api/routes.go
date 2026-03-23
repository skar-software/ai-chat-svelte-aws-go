package api

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v3"

	"aichat_go/internal/chat"
	"aichat_go/internal/config"
	"aichat_go/internal/providers/openai"
	"aichat_go/internal/store"
)

func getModelFromEnv() string {
	if m := os.Getenv("OPENAI_MODEL"); m != "" {
		return m
	}
	return "gpt-4o-mini"
}

func RegisterV1(v1 fiber.Router, cfg *config.Config, logger *slog.Logger) {
	v1.Use(TenantFromHeader())

	memStore := store.NewInMemoryStore()
	provider := openai.NewProvider(cfg.OpenAIAPIKey, getModelFromEnv())
	chatSvc := chat.NewChatService(memStore, provider, logger)

	v1.Post("/responses:stream", StreamChat(cfg, memStore, chatSvc))
	v1.Get("/conversations", ListConversations(memStore))
	v1.Get("/conversations/:id/messages", GetConversationMessages(memStore))
}
