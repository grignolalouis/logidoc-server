// Package response defines the HTTP response types for the API.
package response

import (
	"github.com/logidoc/logidoc-server/internal/core/domain"
)

type Document struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	PageCount   int     `json:"page_count"`
	NodeCount   int     `json:"node_count"`
	Error       *string `json:"error,omitempty"`
	CreatedAt   string  `json:"created_at"`
	IndexedAt   *string `json:"indexed_at,omitempty"`
}

func FromDocument(doc *domain.Document) Document {
	resp := Document{
		ID:          doc.ID,
		Name:        doc.Name,
		Description: doc.Description,
		Status:      string(doc.Status),
		PageCount:   doc.PageCount,
		NodeCount:   doc.NodeCount,
		CreatedAt:   doc.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if doc.Error != "" {
		resp.Error = &doc.Error
	}
	if doc.IndexedAt != nil {
		s := doc.IndexedAt.Format("2006-01-02T15:04:05Z")
		resp.IndexedAt = &s
	}
	return resp
}

type DocumentList struct {
	Documents []Document `json:"documents"`
}

type TOC struct {
	TOC []Node `json:"toc"`
}

type Sections struct {
	Sections []Node `json:"sections"`
}

type Node struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Text      string `json:"text,omitempty"`
	StartPage int    `json:"start_page,omitempty"`
	EndPage   int    `json:"end_page,omitempty"`
	Children  []Node `json:"children"`
}

func FromNodes(nodes []domain.Node) []Node {
	result := make([]Node, len(nodes))
	for i, n := range nodes {
		result[i] = Node{
			ID:        n.ID,
			Title:     n.Title,
			Summary:   n.Summary,
			Text:      n.Text,
			StartPage: n.StartPage,
			EndPage:   n.EndPage,
			Children:  FromNodes(n.Children),
		}
	}
	return result
}

type Error struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
