package index

import "github.com/logidoc/logidoc-server/internal/core/domain"

// FlatSection is a single section detected by the LLM (flat, not nested).
type FlatSection struct {
	Title     string `json:"title"`
	ID        string `json:"id"`
	Summary   string `json:"summary"`
	Level     int    `json:"level"` // 1=chapter, 2=section, 3=subsection
	StartPage int    `json:"start_page"`
}

type treeNode struct {
	FlatSection
	EndPage  int
	Children []treeNode
}

// BuildTree converts a flat section list into a hierarchical tree.
// Computes end_page for each section from the next section's start_page.
func BuildTree(sections []FlatSection, totalPages int) []treeNode {
	if len(sections) == 0 {
		return nil
	}

	type item struct {
		FlatSection
		EndPage int
	}
	items := make([]item, len(sections))
	for i, s := range sections {
		items[i].FlatSection = s
		if i+1 < len(sections) {
			items[i].EndPage = sections[i+1].StartPage - 1
			if items[i].EndPage < items[i].StartPage {
				items[i].EndPage = items[i].StartPage
			}
		} else {
			items[i].EndPage = totalPages
		}
	}

	var roots []treeNode
	for _, it := range items {
		node := treeNode{FlatSection: it.FlatSection, EndPage: it.EndPage}

		inserted := false
		for j := len(roots) - 1; j >= 0; j-- {
			if roots[j].Level < node.Level {
				roots[j].Children = append(roots[j].Children, node)
				inserted = true
				break
			}
			if insertChild(&roots[j], node) {
				inserted = true
				break
			}
		}
		if !inserted {
			roots = append(roots, node)
		}
	}

	return roots
}

func insertChild(parent *treeNode, node treeNode) bool {
	for i := len(parent.Children) - 1; i >= 0; i-- {
		if parent.Children[i].Level < node.Level {
			parent.Children[i].Children = append(parent.Children[i].Children, node)
			return true
		}
		if insertChild(&parent.Children[i], node) {
			return true
		}
	}
	return false
}

// FillText converts tree nodes into domain nodes, populating Text from parsed pages.
func FillText(nodes []treeNode, pages *Pages) []domain.Node {
	result := make([]domain.Node, len(nodes))
	for i, n := range nodes {
		result[i] = domain.Node{
			ID:        n.ID,
			Title:     n.Title,
			Summary:   n.Summary,
			Text:      pages.Read(n.StartPage, n.EndPage),
			StartPage: n.StartPage,
			EndPage:   n.EndPage,
			Children:  FillText(n.Children, pages),
		}
	}
	return result
}

// FillTextEnriched uses ReadEnriched (text + tables + image descriptions).
func FillTextEnriched(nodes []treeNode, pages *Pages) []domain.Node {
	result := make([]domain.Node, len(nodes))
	for i, n := range nodes {
		result[i] = domain.Node{
			ID:        n.ID,
			Title:     n.Title,
			Summary:   n.Summary,
			Text:      pages.ReadEnriched(n.StartPage, n.EndPage),
			StartPage: n.StartPage,
			EndPage:   n.EndPage,
			Children:  FillTextEnriched(n.Children, pages),
		}
	}
	return result
}

// CountNodes returns the total number of nodes in a tree.
func CountNodes(nodes []domain.Node) int {
	c := len(nodes)
	for _, n := range nodes {
		c += CountNodes(n.Children)
	}
	return c
}
