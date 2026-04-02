package handler

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/logidoc/logidoc-server/internal/adapter/primary/http/response"
	"github.com/logidoc/logidoc-server/internal/core/port"
)

type RetrievalHandler struct {
	svc port.RetrievalService
}

func NewRetrievalHandler(svc port.RetrievalService) *RetrievalHandler {
	return &RetrievalHandler{svc: svc}
}

func (h *RetrievalHandler) GetTOC(c fiber.Ctx) error {
	nodes, err := h.svc.GetTOC(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return c.Status(200).JSON(response.TOC{TOC: response.FromNodes(nodes)})
}

func (h *RetrievalHandler) GetSections(c fiber.Ctx) error {
	idsParam := c.Query("ids")
	if idsParam == "" {
		return c.Status(400).JSON(response.Error{Error: "ids query parameter is required"})
	}
	nodeIDs := splitAndTrim(idsParam)

	nodes, err := h.svc.GetSections(c.Context(), c.Params("id"), nodeIDs)
	if err != nil {
		return err
	}
	return c.Status(200).JSON(response.Sections{Sections: response.FromNodes(nodes)})
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
