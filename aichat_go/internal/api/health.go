package api

import "github.com/gofiber/fiber/v3"

func Health(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"ok": true,
		// Lets you confirm the running binary includes the live OpenAI stream path (no canned SSE "part" events).
		"chat_stream": "openai",
	})
}
