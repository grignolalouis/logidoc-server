package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/logidoc/logidoc-server/internal/core/port"
)

type HealthHandler struct {
	docRepo port.DocumentRepository
	version string
}

func NewHealthHandler(docRepo port.DocumentRepository, version string) *HealthHandler {
	return &HealthHandler{docRepo: docRepo, version: version}
}

func (h *HealthHandler) Health(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	mongo := "ok"
	if _, err := h.docRepo.FindAll(ctx); err != nil {
		mongo = "error: " + err.Error()
	}

	status := "ok"
	if mongo != "ok" {
		status = "degraded"
	}

	return c.JSON(fiber.Map{
		"status":  status,
		"version": h.version,
		"mongo":   mongo,
	})
}

func (h *HealthHandler) Version(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"version": h.version})
}
