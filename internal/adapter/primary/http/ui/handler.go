// Package ui provides HTMX-powered server-rendered UI handlers.
package ui

import (
	"bytes"
	"embed"
	"html/template"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/logidoc/logidoc-server/internal/core/domain"
	"github.com/logidoc/logidoc-server/internal/core/port"
)

//go:embed templates/*.html
var templateFS embed.FS

var tmpl = template.Must(template.ParseFS(templateFS, "templates/*.html"))

// Handler serves the UI pages.
type Handler struct {
	docSvc port.DocumentService
	retSvc port.RetrievalService
}

// NewHandler creates a new UI handler.
func NewHandler(docSvc port.DocumentService, retSvc port.RetrievalService) *Handler {
	return &Handler{docSvc: docSvc, retSvc: retSvc}
}

// Dashboard renders the document list page.
func (h *Handler) Dashboard(c fiber.Ctx) error {
	docs, err := h.docSvc.ListAll(c.Context())
	if err != nil {
		return err
	}
	return h.renderPage(c, "dashboard.html", "Documents", fiber.Map{
		"Documents": docs,
	})
}

// DocListPartial returns just the table rows for HTMX polling (no layout).
func (h *Handler) DocListPartial(c fiber.Ctx) error {
	docs, err := h.docSvc.ListAll(c.Context())
	if err != nil {
		return err
	}
	return h.renderFragment(c, "doc-rows", fiber.Map{
		"Documents": docs,
	})
}

// UploadSubmit handles file upload then returns updated table rows.
func (h *Handler) UploadSubmit(c fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).SendString("No file selected")
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(500).SendString("Failed to read file")
	}
	defer f.Close()

	if _, err := h.docSvc.Upload(c.Context(), file.Filename, f); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	// Return updated rows so the table refreshes
	return h.DocListPartial(c)
}

// IndexSubmit triggers indexation then returns updated rows.
func (h *Handler) IndexSubmit(c fiber.Ctx) error {
	if err := h.docSvc.Index(c.Context(), c.Params("id")); err != nil {
		return h.renderFragment(c, "upload-error", fiber.Map{"Error": err.Error()})
	}
	return h.DocListPartial(c)
}

// Document renders the document detail page with TOC.
func (h *Handler) Document(c fiber.Ctx) error {
	doc, err := h.docSvc.Get(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}

	var toc []tocNode
	if doc.Status == domain.StatusReady {
		nodes, err := h.retSvc.GetFullTree(c.Context(), doc.ID)
		if err == nil {
			toc = convertNodes(nodes)
		}
	}

	isPDF := strings.HasSuffix(strings.ToLower(doc.Name), ".pdf")

	return h.renderPage(c, "document.html", doc.Name, fiber.Map{
		"Doc":   doc,
		"TOC":   toc,
		"IsPDF": isPDF,
	})
}

// ---- template helpers ----

type tocNode struct {
	ID        string
	Title     string
	Summary   string
	Text      string
	StartPage int
	EndPage   int
	Children  []tocNode
}

func convertNodes(nodes []domain.Node) []tocNode {
	result := make([]tocNode, len(nodes))
	for i, n := range nodes {
		result[i] = tocNode{
			ID:        n.ID,
			Title:     n.Title,
			Summary:   n.Summary,
			Text:      n.Text,
			StartPage: n.StartPage,
			EndPage:   n.EndPage,
			Children:  convertNodes(n.Children),
		}
	}
	return result
}

// renderPage renders a template wrapped in layout.html.
func (h *Handler) renderPage(c fiber.Ctx, page string, title string, data fiber.Map) error {
	var content bytes.Buffer
	if err := tmpl.ExecuteTemplate(&content, page, data); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout.html", fiber.Map{
		"Title":   title,
		"Content": template.HTML(content.String()),
	}); err != nil {
		return err
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(buf.Bytes())
}

// renderFragment renders a template fragment (no layout, for HTMX swaps).
func (h *Handler) renderFragment(c fiber.Ctx, name string, data any) error {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(buf.Bytes())
}
