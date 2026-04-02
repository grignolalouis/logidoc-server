package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	app "github.com/logidoc/logidoc-server/internal"
	"github.com/logidoc/logidoc-server/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	a, err := app.NewApp(cfg)
	if err != nil {
		slog.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error { return a.HTTPServer.Start() })
	g.Go(func() error { return a.MCPServer.Start() })

	slog.Info("logidoc server started",
		"http", cfg.HTTP.Addr,
		"mcp_transport", cfg.MCP.Transport,
		"mcp_addr", cfg.MCP.Addr,
		"llm_provider", cfg.LLM.Provider,
		"llm_model", cfg.LLM.Model,
	)

	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
}
