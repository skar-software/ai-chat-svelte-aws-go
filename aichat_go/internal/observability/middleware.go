package observability

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)
		reqID := requestid.FromContext(c)
		tenantID := c.Locals("tenant_id")
		userID := c.Locals("user_id")
		args := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency_ms", latency.Milliseconds(),
		}
		if reqID != "" {
			args = append(args, "request_id", reqID)
		}
		if tenantID != nil {
			args = append(args, "tenant_id", tenantID)
		}
		if userID != nil {
			args = append(args, "user_id", userID)
		}
		logger.Info("request", args...)
		return err
	}
}
