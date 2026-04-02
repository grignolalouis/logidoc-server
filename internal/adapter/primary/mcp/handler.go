package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcplib "trpc.group/trpc-go/trpc-mcp-go"

	"github.com/logidoc/logidoc-server/internal/core/port"
)

type handler struct {
	docSvc port.DocumentService
	retSvc port.RetrievalService
}

func (h *handler) listDocuments(ctx context.Context, _ *mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	docs, err := h.docSvc.List(ctx)
	if err != nil {
		return nil, err
	}
	return mcplib.NewTextResult(formatDocumentList(docs)), nil
}

func (h *handler) getTOC(ctx context.Context, req *mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	docID, ok := req.Params.Arguments["doc_id"].(string)
	if !ok || docID == "" {
		return nil, fmt.Errorf("doc_id is required")
	}
	toc, err := h.retSvc.GetTOC(ctx, docID)
	if err != nil {
		return nil, err
	}
	return mcplib.NewTextResult(formatTOC(toc, 0)), nil
}

func (h *handler) getSections(ctx context.Context, req *mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	docID, ok1 := req.Params.Arguments["doc_id"].(string)
	nodeIDsStr, ok2 := req.Params.Arguments["node_ids"].(string)
	if !ok1 || docID == "" || !ok2 || nodeIDsStr == "" {
		return nil, fmt.Errorf("doc_id and node_ids are required")
	}

	var nodeIDs []string
	for _, id := range strings.Split(nodeIDsStr, ",") {
		if t := strings.TrimSpace(id); t != "" {
			nodeIDs = append(nodeIDs, t)
		}
	}

	nodes, err := h.retSvc.GetSections(ctx, docID, nodeIDs)
	if err != nil {
		return nil, err
	}

	b, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal sections: %w", err)
	}
	return mcplib.NewTextResult(string(b)), nil
}
