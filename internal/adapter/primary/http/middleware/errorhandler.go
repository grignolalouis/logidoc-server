package middleware

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/logidoc/logidoc-server/internal/core/domain"
)

func ErrorHandler(logger *slog.Logger) func(c fiber.Ctx, err error) error {
	return func(c fiber.Ctx, err error) error {
		code := 500
		message := "internal server error"

		var notFound *domain.NotFoundError
		var notReady *domain.NotReadyError
		var validation *domain.ValidationError
		var fiberErr *fiber.Error

		switch {
		case errors.As(err, &notFound):
			code = 404
			message = notFound.Error()
		case errors.As(err, &notReady):
			code = 409
			message = notReady.Error()
		case errors.As(err, &validation):
			code = 400
			message = validation.Message
		case errors.As(err, &fiberErr):
			code = fiberErr.Code
			message = fiberErr.Message
		}

		if code >= 500 {
			logger.Error("request error", "path", c.Path(), "method", c.Method(), "error", err)
		}

		return c.Status(code).JSON(fiber.Map{
			"error":   message,
			"message": err.Error(),
		})
	}
}
