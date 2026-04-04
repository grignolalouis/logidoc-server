package handler

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/logidoc/logidoc-server/internal/adapter/primary/http/response"
	"github.com/logidoc/logidoc-server/internal/core/port"
)

type Document struct {
	svc       port.DocumentService
	fileStore port.FileRepository
}

func NewDocument(svc port.DocumentService, fileStore port.FileRepository) *Document {
	return &Document{svc: svc, fileStore: fileStore}
}

func (h *Document) Upload(c fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(response.Error{Error: "file is required"})
	}
	f, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(response.Error{Error: "failed to open file"})
	}
	defer f.Close()

	doc, err := h.svc.Upload(c.Context(), file.Filename, f)
	if err != nil {
		return c.Status(500).JSON(response.Error{Error: err.Error()})
	}
	return c.Status(201).JSON(response.FromDocument(doc))
}

func (h *Document) List(c fiber.Ctx) error {
	docs, err := h.svc.List(c.Context())
	if err != nil {
		return c.Status(500).JSON(response.Error{Error: err.Error()})
	}
	items := make([]response.Document, len(docs))
	for i := range docs {
		items[i] = response.FromDocument(&docs[i])
	}
	return c.Status(200).JSON(response.DocumentList{Documents: items})
}

func (h *Document) Get(c fiber.Ctx) error {
	doc, err := h.svc.Get(c.Context(), c.Params("id"))
	if err != nil {
		return err // ErrorHandler maps domain errors to HTTP status
	}
	return c.Status(200).JSON(response.FromDocument(doc))
}

func (h *Document) File(c fiber.Ctx) error {
	id := c.Params("id")
	doc, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return err
	}
	data, err := h.fileStore.Load(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(response.Error{Error: "file not found"})
	}
	// Sanitize filename to prevent header injection and path traversal
	safeName := filepath.Base(doc.Name)
	if strings.HasSuffix(strings.ToLower(safeName), ".pdf") {
		c.Set("Content-Type", "application/pdf")
	} else {
		c.Set("Content-Type", "text/plain; charset=utf-8")
	}
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename*=UTF-8''%s", url.QueryEscape(safeName)))
	return c.Send(data)
}

func (h *Document) Index(c fiber.Ctx) error {
	if err := h.svc.Index(c.Context(), c.Params("id")); err != nil {
		return c.Status(400).JSON(response.Error{Error: err.Error()})
	}
	return c.Status(202).JSON(fiber.Map{"status": "indexing"})
}

func (h *Document) Delete(c fiber.Ctx) error {
	if err := h.svc.Delete(c.Context(), c.Params("id")); err != nil {
		return err // ErrorHandler maps domain errors
	}
	return c.SendStatus(204)
}
