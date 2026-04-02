package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func Recovery(logger *slog.Logger) fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, e any) {
			logger.Error("panic recovered", "path", c.Path(), "method", c.Method(), "error", e)
		},
	})
}
