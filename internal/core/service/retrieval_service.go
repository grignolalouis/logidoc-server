package service

import (
	"context"

	"github.com/logidoc/logidoc-server/internal/core/domain"
	"github.com/logidoc/logidoc-server/internal/core/port"
)

var _ port.RetrievalService = (*RetrievalService)(nil)

// RetrievalService provides read access to document indexes.
type RetrievalService struct {
	indexRepo port.IndexRepository
}

// NewRetrievalService creates a new RetrievalService.
func NewRetrievalService(indexRepo port.IndexRepository) *RetrievalService {
	return &RetrievalService{indexRepo: indexRepo}
}

// GetTOC returns the compact tree (titles + summaries, no full text) for a document.
func (s *RetrievalService) GetTOC(ctx context.Context, docID string) ([]domain.Node, error) {
	index, err := s.indexRepo.FindByDocID(ctx, docID)
	if err != nil {
		return nil, err
	}
	return stripText(index.Tree), nil
}

// GetFullTree returns the complete tree with text and page numbers.
func (s *RetrievalService) GetFullTree(ctx context.Context, docID string) ([]domain.Node, error) {
	index, err := s.indexRepo.FindByDocID(ctx, docID)
	if err != nil {
		return nil, err
	}
	return index.Tree, nil
}

// GetSections returns the full nodes (with text) for the given node IDs.
func (s *RetrievalService) GetSections(ctx context.Context, docID string, nodeIDs []string) ([]domain.Node, error) {
	index, err := s.indexRepo.FindByDocID(ctx, docID)
	if err != nil {
		return nil, err
	}
	return extractNodes(index.Tree, nodeIDs), nil
}

// stripText returns a copy of the tree with Text fields emptied (TOC only).
func stripText(nodes []domain.Node) []domain.Node {
	result := make([]domain.Node, len(nodes))
	for i, n := range nodes {
		result[i] = domain.Node{
			ID:        n.ID,
			Title:     n.Title,
			Summary:   n.Summary,
			StartPage: n.StartPage,
			EndPage:   n.EndPage,
			Children:  stripText(n.Children),
		}
	}
	return result
}

// extractNodes finds nodes by ID in a tree (recursive).
func extractNodes(tree []domain.Node, ids []string) []domain.Node {
	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	var result []domain.Node
	var walk func(nodes []domain.Node)
	walk = func(nodes []domain.Node) {
		for _, n := range nodes {
			if _, ok := idSet[n.ID]; ok {
				result = append(result, n)
			}
			walk(n.Children)
		}
	}
	walk(tree)
	return result
}
