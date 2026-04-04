package mcp

import (
	"context"
	"log/slog"

	mcplib "trpc.group/trpc-go/trpc-mcp-go"

	"github.com/logidoc/logidoc-server/internal/config"
	"github.com/logidoc/logidoc-server/internal/core/port"
)

type Server struct {
	stdio  *mcplib.StdioServer
	http   *mcplib.Server
	cfg    config.MCPConfig
	logger *slog.Logger
}

func NewServer(cfg config.MCPConfig, docSvc port.DocumentService, retSvc port.RetrievalService, logger *slog.Logger) *Server {
	s := &Server{cfg: cfg, logger: logger}
	h := &handler{docSvc: docSvc, retSvc: retSvc}
	defs := tools()

	if cfg.Stdio {
		s.stdio = mcplib.NewStdioServer("logidoc", "1.0.0")
		s.stdio.RegisterTool(defs[0], h.listDocuments)
		s.stdio.RegisterTool(defs[1], h.search)
		s.stdio.RegisterTool(defs[2], h.getTOC)
		s.stdio.RegisterTool(defs[3], h.getSections)
	} else {
		s.http = mcplib.NewServer("logidoc", "1.0.0", mcplib.WithServerAddress(cfg.Addr))
		s.http.RegisterTool(defs[0], h.listDocuments)
		s.http.RegisterTool(defs[1], h.search)
		s.http.RegisterTool(defs[2], h.getTOC)
		s.http.RegisterTool(defs[3], h.getSections)
	}
	return s
}

func (s *Server) Start() error {
	if s.cfg.Stdio {
		s.logger.Info("starting MCP server", "transport", "stdio")
		return s.stdio.Start()
	}
	s.logger.Info("starting MCP server", "transport", s.cfg.Transport, "addr", s.cfg.Addr)
	return s.http.Start()
}

func (s *Server) Shutdown(_ context.Context) error {
	return nil
}
