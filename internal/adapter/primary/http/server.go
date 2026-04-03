package http

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/logidoc/logidoc-server/internal/adapter/primary/http/middleware"
	"github.com/logidoc/logidoc-server/internal/adapter/primary/http/router"
	"github.com/logidoc/logidoc-server/internal/config"
	"github.com/logidoc/logidoc-server/internal/core/port"
)

type Server struct {
	app    *fiber.App
	cfg    config.HTTPConfig
	logger *slog.Logger
}

func NewServer(cfg config.Config, docSvc port.DocumentService, retSvc port.RetrievalService, fileStore port.FileStore, docRepo port.DocumentRepository, logger *slog.Logger) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		BodyLimit:    cfg.HTTP.BodyLimitMB * 1024 * 1024,
		ErrorHandler: middleware.ErrorHandler(logger),
	})

	router.Setup(app, cfg, docSvc, retSvc, fileStore, docRepo, logger)

	return &Server{app: app, cfg: cfg.HTTP, logger: logger}
}

func (s *Server) Start() error {
	s.logger.Info("starting HTTP server", "addr", s.cfg.Addr)
	return s.app.Listen(s.cfg.Addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
