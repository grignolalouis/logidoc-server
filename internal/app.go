package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	httpapi "github.com/logidoc/logidoc-server/internal/adapter/primary/http"
	mcpapi "github.com/logidoc/logidoc-server/internal/adapter/primary/mcp"
	"github.com/logidoc/logidoc-server/internal/adapter/secondary/llm"
	mongorepo "github.com/logidoc/logidoc-server/internal/adapter/secondary/repository/mongo"
	"github.com/logidoc/logidoc-server/internal/config"
	"github.com/logidoc/logidoc-server/internal/core/service"
	"github.com/logidoc/logidoc-server/internal/core/service/indexer"
	infralog "github.com/logidoc/logidoc-server/internal/infrastructure/logger"
)

// App holds all the server components.
type App struct {
	HTTPServer *httpapi.Server
	MCPServer  *mcpapi.Server
	MongoDB    *mongorepo.Connection
	Logger     *slog.Logger
}

// NewApp creates and wires all dependencies.
func NewApp(cfg *config.Config) (*App, error) {
	logger := infralog.New(cfg.Logger)
	slog.SetDefault(logger)

	// Secondary adapters
	mongoConn, err := mongorepo.NewConnection(cfg.Mongo)
	if err != nil {
		return nil, fmt.Errorf("mongo connection: %w", err)
	}
	docRepo := mongorepo.NewDocumentRepo(mongoConn)
	indexRepo := mongorepo.NewIndexRepo(mongoConn)
	fileStore := mongorepo.NewFileStore(mongoConn)

	// LLM provider
	llmModel, err := llm.NewModel(cfg.LLM, logger)
	if err != nil {
		return nil, fmt.Errorf("llm provider: %w", err)
	}

	// Core services
	indexerSvc := indexer.NewService(llmModel, docRepo, indexRepo, logger)
	docSvc := service.NewDocumentService(docRepo, indexRepo, fileStore, indexerSvc, logger)
	retSvc := service.NewRetrievalService(indexRepo)

	// Primary adapters
	httpServer := httpapi.NewServer(cfg.HTTP, docSvc, retSvc, fileStore, logger)
	mcpServer := mcpapi.NewServer(cfg.MCP, docSvc, retSvc, logger)

	return &App{
		HTTPServer: httpServer,
		MCPServer:  mcpServer,
		MongoDB:    mongoConn,
		Logger:     logger,
	}, nil
}

// Shutdown gracefully stops all components.
func (a *App) Shutdown(ctx context.Context) error {
	a.Logger.Info("shutting down services...")
	var errs []error
	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("http shutdown: %w", err))
	}
	if err := a.MCPServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("mcp shutdown: %w", err))
	}
	if err := a.MongoDB.Close(ctx); err != nil {
		errs = append(errs, fmt.Errorf("mongo close: %w", err))
	}
	return errors.Join(errs...)
}
