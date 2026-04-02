package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

// APIKeyAuth returns a middleware that validates the API key from the Authorization header.
// If apiKey is empty, the middleware is a no-op (auth disabled).
func APIKeyAuth(apiKey string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if apiKey == "" {
			return c.Next()
		}

		// Skip auth for health, version, and UI routes
		path := c.Path()
		if path == "/health" || path == "/version" || strings.HasPrefix(path, "/ui") {
			return c.Next()
		}

		header := c.Get("Authorization")
		if header == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing Authorization header"})
		}

		token := strings.TrimPrefix(header, "Bearer ")
		if token != apiKey {
			return c.Status(401).JSON(fiber.Map{"error": "invalid API key"})
		}

		return c.Next()
	}
}
