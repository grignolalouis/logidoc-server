package middleware

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/logidoc/logidoc-server/internal/core/domain"
)

// ErrorHandler returns a Fiber error handler that maps domain errors to HTTP status codes
// and logs unexpected errors.
func ErrorHandler(logger *slog.Logger) func(c fiber.Ctx, err error) error {
	return func(c fiber.Ctx, err error) error {
		code := 500
		message := "internal server error"

		// Map domain errors to HTTP codes
		switch {
		case errors.Is(err, domain.ErrDocumentNotFound):
			code = 404
			message = "document not found"
		case errors.Is(err, domain.ErrIndexNotFound):
			code = 404
			message = "index not found"
		case errors.Is(err, domain.ErrDocumentNotReady):
			code = 409
			message = "document is not yet indexed"
		case errors.Is(err, domain.ErrInvalidDocument):
			code = 400
			message = "invalid document"
		case errors.Is(err, domain.ErrIndexingFailed):
			code = 500
			message = "indexing failed"
		default:
			// Check Fiber errors (404 from router, etc.)
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				code = fiberErr.Code
				message = fiberErr.Message
			}
		}

		if code >= 500 {
			logger.Error("request error",
				"path", c.Path(),
				"method", c.Method(),
				"status", code,
				"error", err,
			)
		}

		return c.Status(code).JSON(fiber.Map{
			"error":   message,
			"message": err.Error(),
		})
	}
}
