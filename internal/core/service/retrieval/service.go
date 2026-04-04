package retrieval

import (
	"context"
	"sort"
	"strings"

	"github.com/logidoc/logidoc-server/internal/core/domain"
	"github.com/logidoc/logidoc-server/internal/core/port"
)

type RetrievalService struct {
	docRepo   port.DocumentRepository
	indexRepo port.IndexRepository
}

func NewRetrievalService(docRepo port.DocumentRepository, indexRepo port.IndexRepository) *RetrievalService {
	return &RetrievalService{docRepo: docRepo, indexRepo: indexRepo}
}

func (s *RetrievalService) GetTOC(ctx context.Context, docID string) ([]domain.Node, error) {
	index, err := s.indexRepo.FindByDocID(ctx, docID)
	if err != nil {
		return nil, err
	}
	return stripText(index.Tree), nil
}

func (s *RetrievalService) GetFullTree(ctx context.Context, docID string) ([]domain.Node, error) {
	index, err := s.indexRepo.FindByDocID(ctx, docID)
	if err != nil {
		return nil, err
	}
	return index.Tree, nil
}

func (s *RetrievalService) GetSections(ctx context.Context, docID string, nodeIDs []string) ([]domain.Node, error) {
	index, err := s.indexRepo.FindByDocID(ctx, docID)
	if err != nil {
		return nil, err
	}
	return extractNodes(index.Tree, nodeIDs), nil
}

func (s *RetrievalService) Search(ctx context.Context, query string) ([]domain.SearchHit, error) {
	docs, err := s.docRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	queryWords := tokenize(query)
	var hits []domain.SearchHit

	for _, doc := range docs {
		if doc.Status != domain.StatusReady {
			continue
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		index, err := s.indexRepo.FindByDocID(ctx, doc.ID)
		if err != nil {
			continue
		}

		walkNodes(index.Tree, func(node domain.Node) {
			score := matchScore(queryWords, node.Title, node.Summary)
			if score > 0 {
				hits = append(hits, domain.SearchHit{
					DocID:     doc.ID,
					DocName:   doc.Name,
					NodeID:    node.ID,
					NodeTitle: node.Title,
					Summary:   node.Summary,
					StartPage: node.StartPage,
					EndPage:   node.EndPage,
					Score:     score,
				})
			}
		})
	}

	sort.Slice(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if len(hits) > 10 {
		hits = hits[:10]
	}
	return hits, nil
}

func tokenize(s string) []string {
	words := strings.Fields(strings.ToLower(s))
	var result []string
	for _, w := range words {
		if len(w) >= 3 {
			result = append(result, w)
		}
	}
	return result
}

func matchScore(queryWords []string, title, summary string) float64 {
	if len(queryWords) == 0 {
		return 0
	}
	target := strings.ToLower(title + " " + summary)
	var matched float64
	for _, w := range queryWords {
		if strings.Contains(target, w) {
			matched++
		}
	}
	return matched / float64(len(queryWords))
}

func walkNodes(nodes []domain.Node, fn func(domain.Node)) {
	for _, n := range nodes {
		fn(n)
		walkNodes(n.Children, fn)
	}
}

func stripText(nodes []domain.Node) []domain.Node {
	result := make([]domain.Node, len(nodes))
	for i, n := range nodes {
		result[i] = domain.Node{
			ID: n.ID, Title: n.Title, Summary: n.Summary,
			StartPage: n.StartPage, EndPage: n.EndPage,
			Children: stripText(n.Children),
		}
	}
	return result
}

func extractNodes(tree []domain.Node, ids []string) []domain.Node {
	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	var result []domain.Node
	walkNodes(tree, func(n domain.Node) {
		if _, ok := idSet[n.ID]; ok {
			result = append(result, n)
		}
	})
	return result
}
