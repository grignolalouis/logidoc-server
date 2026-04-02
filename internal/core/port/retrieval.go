package port

import (
	"context"

	"github.com/logidoc/logidoc-server/internal/core/domain"
)

// RetrievalService defines the use cases for document retrieval.
type RetrievalService interface {
	GetTOC(ctx context.Context, docID string) ([]domain.Node, error)      // titles + summaries only (for MCP/agents)
	GetFullTree(ctx context.Context, docID string) ([]domain.Node, error) // full tree with text + pages (for UI)
	GetSections(ctx context.Context, docID string, nodeIDs []string) ([]domain.Node, error)
}
