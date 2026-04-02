package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Logger(logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		logger.Info("request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", time.Since(start).Round(time.Microsecond),
			"ip", c.IP(),
		)
		return err
	}
}
