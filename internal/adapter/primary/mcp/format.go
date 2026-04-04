package mcp

import (
	"fmt"
	"strings"

	"github.com/logidoc/logidoc-server/internal/core/domain"
)

func formatTOC(nodes []domain.Node, depth int) string {
	var sb strings.Builder
	indent := strings.Repeat("  ", depth)
	for _, n := range nodes {
		fmt.Fprintf(&sb, "%s- [%s] %s\n", indent, n.ID, n.Title)
		if n.Summary != "" {
			fmt.Fprintf(&sb, "%s  %s\n", indent, n.Summary)
		}
		if len(n.Children) > 0 {
			sb.WriteString(formatTOC(n.Children, depth+1))
		}
	}
	return sb.String()
}

func formatDocumentList(docs []domain.Document) string {
	if len(docs) == 0 {
		return "No documents indexed."
	}
	var sb strings.Builder
	for _, doc := range docs {
		fmt.Fprintf(&sb, "- **%s** [%s] (%s)\n", doc.Name, doc.ID, doc.Status)
		if doc.Description != "" {
			fmt.Fprintf(&sb, "  %s\n", doc.Description)
		}
	}
	return sb.String()
}

func formatSearchHits(hits []domain.SearchHit) string {
	if len(hits) == 0 {
		return "No results found."
	}
	var sb strings.Builder
	for _, h := range hits {
		fmt.Fprintf(&sb, "- **%s** > %s", h.DocName, h.NodeTitle)
		if h.StartPage > 0 {
			fmt.Fprintf(&sb, " (p.%d-%d)", h.StartPage, h.EndPage)
		}
		sb.WriteString("\n")
		fmt.Fprintf(&sb, "  Doc: %s | Node: %s\n", h.DocID, h.NodeID)
		if h.Summary != "" {
			fmt.Fprintf(&sb, "  %s\n", h.Summary)
		}
	}
	return sb.String()
}
