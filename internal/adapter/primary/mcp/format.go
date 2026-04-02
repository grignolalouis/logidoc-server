package mcp

import (
	"fmt"
	"strings"

	"github.com/logidoc/logidoc-server/internal/core/domain"
)

// formatTOC renders the tree as indented text readable by an LLM agent.
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

// formatDocumentList renders documents as a markdown list for an LLM agent.
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
