package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

type HealthChecker interface {
	Check(ctx context.Context) error
}

type Health struct {
	checker HealthChecker
	version string
}

func NewHealth(checker HealthChecker, version string) *Health {
	return &Health{checker: checker, version: version}
}

func (h *Health) HealthCheck(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	mongo := "ok"
	if err := h.checker.Check(ctx); err != nil {
		mongo = "error"
	}

	status := "ok"
	if mongo != "ok" {
		status = "degraded"
	}

	return c.JSON(fiber.Map{"status": status, "version": h.version, "mongo": mongo})
}

func (h *Health) Version(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"version": h.version})
}
