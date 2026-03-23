package api

import (
	"github.com/gofiber/fiber/v3"
)

// TenantFromHeader sets tenant_id and workspace_id in Locals from headers (for development).
// In production, derive these from the authenticated session/JWT instead of trusting headers.
func TenantFromHeader() fiber.Handler {
	return func(c fiber.Ctx) error {
		if v := c.Get("X-Tenant-ID"); v != "" {
			c.Locals("tenant_id", v)
		}
		if v := c.Get("X-Workspace-ID"); v != "" {
			c.Locals("workspace_id", v)
		}
		if v := c.Get("Authorization"); v != "" && len(v) > 7 && v[:7] == "Bearer " {
			c.Locals("user_id", v[7:])
		}
		return c.Next()
	}
}
