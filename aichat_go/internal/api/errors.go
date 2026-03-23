package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

func ErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		logger.Error("handler error", "error", err, "path", c.Path())
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}
		return c.Status(code).JSON(fiber.Map{"error": err.Error()})
	}
}
