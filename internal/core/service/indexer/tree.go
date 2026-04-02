package indexer

import (
	"strings"

	"github.com/logidoc/logidoc-server/internal/core/domain"
)

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

// CalibratePages detects the offset between logical page numbers (from the TOC)
// and physical PDF pages. Academic PDFs often have title pages, TOC pages, etc.
// before page "1" of the actual content.
//
// Strategy: take the first few section titles, search for them in the extracted pages,
// and compute the most common offset.
func CalibratePages(sections []FlatSection, pages *Pages) []FlatSection {
	if len(sections) == 0 || pages.Count() == 0 {
		return sections
	}

	// Already calibrated? Check if first section's page has its title
	first := sections[0]
	if first.StartPage >= 1 && first.StartPage <= pages.Count() {
		pageText := strings.ToLower(pages.Content[first.StartPage-1])
		titleWords := strings.Fields(strings.ToLower(first.Title))
		if len(titleWords) > 0 && strings.Contains(pageText, titleWords[len(titleWords)-1]) {
			return sections // no offset needed
		}
	}

	// Try offsets from 0 to 20 — find the one where most titles match their pages
	bestOffset := 0
	bestScore := 0

	for offset := 0; offset <= 20 && offset < pages.Count(); offset++ {
		score := 0
		for _, s := range sections {
			physPage := s.StartPage + offset
			if physPage < 1 || physPage > pages.Count() {
				continue
			}
			pageText := strings.ToLower(pages.Content[physPage-1])
			// Check if any significant word from the title appears on this page
			for _, word := range strings.Fields(strings.ToLower(s.Title)) {
				if len(word) >= 4 && strings.Contains(pageText, word) {
					score++
					break
				}
			}
		}
		if score > bestScore {
			bestScore = score
			bestOffset = offset
		}
	}

	if bestOffset == 0 {
		return sections
	}

	// Apply offset
	calibrated := make([]FlatSection, len(sections))
	copy(calibrated, sections)
	for i := range calibrated {
		calibrated[i].StartPage += bestOffset
		if calibrated[i].StartPage > pages.Count() {
			calibrated[i].StartPage = pages.Count()
		}
	}

	return calibrated
}

// BuildTree converts a flat section list into a hierarchical tree.
// Computes end_page for each section from the next section's start_page.
func BuildTree(sections []FlatSection, totalPages int) []treeNode {
	if len(sections) == 0 {
		return nil
	}

	// Compute end_page
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

	// Build tree using levels
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

// CountNodes returns the total number of nodes in a tree.
func CountNodes(nodes []domain.Node) int {
	c := len(nodes)
	for _, n := range nodes {
		c += CountNodes(n.Children)
	}
	return c
}
