// Package router registers all HTTP middleware and routes.
package router

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/logidoc/logidoc-server/internal/adapter/primary/http/handler"
	"github.com/logidoc/logidoc-server/internal/adapter/primary/http/middleware"
	"github.com/logidoc/logidoc-server/internal/adapter/primary/http/ui"
	"github.com/logidoc/logidoc-server/internal/config"
	"github.com/logidoc/logidoc-server/internal/core/port"
)

// Setup registers all middleware and routes on the Fiber app.
func Setup(app *fiber.App, cfg config.HTTPConfig, docSvc port.DocumentService, retSvc port.RetrievalService, fileStore port.FileStore, logger *slog.Logger) {
	registerMiddleware(app, cfg, logger)
	registerAPI(app, docSvc, retSvc, fileStore)
	registerUI(app, docSvc, retSvc)
}

func registerMiddleware(app *fiber.App, cfg config.HTTPConfig, logger *slog.Logger) {
	app.Use(requestid.New())
	app.Use(middleware.Recovery(logger))
	app.Use(middleware.Logger(logger))
	app.Use(middleware.CORS([]string{cfg.CORSOrigins}))
	app.Use(middleware.RateLimit(cfg.RateLimit))
}

func registerAPI(app *fiber.App, docSvc port.DocumentService, retSvc port.RetrievalService, fileStore port.FileStore) {
	docH := handler.NewDocumentHandler(docSvc, fileStore)
	retH := handler.NewRetrievalHandler(retSvc)

	v1 := app.Group("/v1")
	v1.Post("/documents", docH.Upload)
	v1.Get("/documents", docH.List)
	v1.Get("/documents/:id", docH.Get)
	v1.Get("/documents/:id/file", docH.File)
	v1.Post("/documents/:id/index", docH.Index)
	v1.Delete("/documents/:id", docH.Delete)
	v1.Get("/documents/:id/toc", retH.GetTOC)
	v1.Get("/documents/:id/sections", retH.GetSections)
}

func registerUI(app *fiber.App, docSvc port.DocumentService, retSvc port.RetrievalService) {
	h := ui.NewHandler(docSvc, retSvc)

	app.Get("/ui", h.Dashboard)
	app.Post("/ui/upload", h.UploadSubmit)
	app.Get("/ui/documents/:id", h.Document)
	app.Post("/ui/documents/:id/index", h.IndexSubmit)
	app.Get("/ui/partials/doc-list", h.DocListPartial)

	app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().To("/ui")
	})
}
