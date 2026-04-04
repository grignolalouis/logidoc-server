package port

import (
	"context"

	"github.com/logidoc/logidoc-server/internal/core/domain"
)

type RetrievalService interface {
	GetTOC(ctx context.Context, docID string) ([]domain.Node, error)
	GetFullTree(ctx context.Context, docID string) ([]domain.Node, error)
	GetSections(ctx context.Context, docID string, nodeIDs []string) ([]domain.Node, error)
	Search(ctx context.Context, query string) ([]domain.SearchHit, error)
}
