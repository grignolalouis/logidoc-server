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

// Server is the HTTP server wrapping Fiber v3.
type Server struct {
	app    *fiber.App
	cfg    config.HTTPConfig
	logger *slog.Logger
}

// NewServer creates and configures a new HTTP server.
func NewServer(cfg config.HTTPConfig, docSvc port.DocumentService, retSvc port.RetrievalService, fileStore port.FileStore, logger *slog.Logger) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		BodyLimit:    50 * 1024 * 1024, // 50MB
		ErrorHandler: middleware.ErrorHandler(logger),
	})

	router.Setup(app, cfg, docSvc, retSvc, fileStore, logger)

	return &Server{app: app, cfg: cfg, logger: logger}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	s.logger.Info("starting HTTP server", "addr", s.cfg.Addr)
	return s.app.Listen(s.cfg.Addr)
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
