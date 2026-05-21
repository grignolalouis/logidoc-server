package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	agentmodel "trpc.group/trpc-go/trpc-agent-go/model"

	httpapi "github.com/logidoc/logidoc-server/internal/adapter/primary/http"
	mcpapi "github.com/logidoc/logidoc-server/internal/adapter/primary/mcp"
	"github.com/logidoc/logidoc-server/internal/adapter/secondary/llm"
	mongorepo "github.com/logidoc/logidoc-server/internal/adapter/secondary/repository/mongo"
	"github.com/logidoc/logidoc-server/internal/config"
	"github.com/logidoc/logidoc-server/internal/core/service/document"
	"github.com/logidoc/logidoc-server/internal/core/service/index"
	"github.com/logidoc/logidoc-server/internal/core/service/retrieval"
	obslog "github.com/logidoc/logidoc-server/internal/observability/logger"
)

type App struct {
	HTTPServer *httpapi.Server
	MCPServer  *mcpapi.Server
	MongoDB    *mongorepo.Connection
	Logger     *slog.Logger
}

func NewApp(cfg *config.Config) (*App, error) {
	logger := obslog.New(cfg.Logger)
	slog.SetDefault(logger)

	logger.Info("starting logidoc",
		"version", cfg.App.Version,
		"llm_provider", cfg.LLM.Provider,
		"llm_model", cfg.LLM.Model,
		"http", cfg.HTTP.Addr,
		"mcp", cfg.MCP.Addr,
	)

	mongoConn, err := mongorepo.NewConnection(cfg.Mongo)
	if err != nil {
		return nil, fmt.Errorf("mongo connection: %w", err)
	}
	docRepo := mongorepo.NewDocumentRepo(mongoConn)
	indexRepo := mongorepo.NewIndexRepo(mongoConn)
	fileStore := mongorepo.NewFileRepository(mongoConn)

	llmModel, err := llm.NewModel(cfg.LLM, logger)
	if err != nil {
		return nil, fmt.Errorf("llm provider: %w", err)
	}

	var visionModel agentmodel.Model
	if cfg.Vision.Enabled {
		visionModel, err = llm.NewModel(config.LLMConfig{
			Provider: cfg.Vision.Provider,
			Model:    cfg.Vision.Model,
			APIKey:   cfg.Vision.APIKey,
			BaseURL:  cfg.Vision.BaseURL,
		}, logger)
		if err != nil {
			return nil, fmt.Errorf("vision provider: %w", err)
		}
	}

	indexerSvc := index.NewService(llmModel, visionModel, docRepo, indexRepo, index.Config{
		MaxPagesPerNode:        cfg.Indexer.MaxPagesPerNode,
		EnableTableExtraction:  cfg.Indexer.EnableTableExtraction,
		EnableImageDescription: cfg.Indexer.EnableImageDescription,
	}, logger)
	docSvc := document.NewDocumentService(docRepo, indexRepo, fileStore, indexerSvc, logger)
	retSvc := retrieval.NewRetrievalService(docRepo, indexRepo)

	httpServer := httpapi.NewServer(*cfg, docSvc, retSvc, fileStore, mongoConn, logger)
	mcpServer := mcpapi.NewServer(cfg.MCP, docSvc, retSvc, logger)

	return &App{
		HTTPServer: httpServer,
		MCPServer:  mcpServer,
		MongoDB:    mongoConn,
		Logger:     logger,
	}, nil
}

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
