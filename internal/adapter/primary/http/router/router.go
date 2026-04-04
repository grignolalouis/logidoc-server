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

func Setup(app *fiber.App, cfg config.Config, docSvc port.DocumentService, retSvc port.RetrievalService, fileStore port.FileRepository, healthChecker handler.HealthChecker, logger *slog.Logger) {
	app.Use(requestid.New())
	app.Use(middleware.Recovery(logger))
	app.Use(middleware.Logger(logger))
	app.Use(middleware.CORS([]string{cfg.HTTP.CORSOrigins}))
	app.Use(middleware.RateLimit(cfg.HTTP.RateLimit))
	app.Use(middleware.APIKeyAuth(cfg.App.APIKey))

	healthH := handler.NewHealth(healthChecker, cfg.App.Version)
	app.Get("/health", healthH.HealthCheck)
	app.Get("/version", healthH.Version)

	docH := handler.NewDocument(docSvc, fileStore)
	retH := handler.NewRetrieval(retSvc)

	v1 := app.Group("/v1")
	v1.Post("/documents", docH.Upload)
	v1.Get("/documents", docH.List)
	v1.Get("/documents/:id", docH.Get)
	v1.Get("/documents/:id/file", docH.File)
	v1.Post("/documents/:id/index", docH.Index)
	v1.Delete("/documents/:id", docH.Delete)
	v1.Get("/documents/:id/toc", retH.GetTOC)
	v1.Get("/documents/:id/sections", retH.GetSections)
	v1.Get("/search", retH.Search)

	uiH := ui.NewHandler(docSvc, retSvc)
	app.Get("/ui", uiH.Dashboard)
	app.Post("/ui/upload", uiH.UploadSubmit)
	app.Get("/ui/documents/:id", uiH.Document)
	app.Post("/ui/documents/:id/index", uiH.IndexSubmit)
	app.Get("/ui/partials/doc-list", uiH.DocListPartial)

	app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().To("/ui")
	})
}
