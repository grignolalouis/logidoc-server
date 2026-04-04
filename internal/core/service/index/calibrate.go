package index

import (
	"log/slog"
	"strings"
)

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

	first := sections[0]
	if first.StartPage >= 1 && first.StartPage <= pages.Count() {
		pageText := strings.ToLower(pages.Content[first.StartPage-1])
		titleWords := strings.Fields(strings.ToLower(first.Title))
		if len(titleWords) > 0 && strings.Contains(pageText, titleWords[len(titleWords)-1]) {
			return sections
		}
	}

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

// VerifyCalibration checks that each section's title appears on its assigned page.
// If not found, searches ±3 nearby pages and adjusts.
func VerifyCalibration(sections []FlatSection, pages *Pages, logger *slog.Logger) []FlatSection {
	for i, s := range sections {
		if s.StartPage < 1 || s.StartPage > pages.Count() {
			continue
		}

		words := significantWords(s.Title)
		if len(words) == 0 {
			continue
		}

		if containsAny(pages.Content[s.StartPage-1], words) {
			continue
		}

		for offset := 1; offset <= 3; offset++ {
			found := false
			for _, delta := range []int{-offset, offset} {
				candidate := s.StartPage + delta
				if candidate < 1 || candidate > pages.Count() {
					continue
				}
				if containsAny(pages.Content[candidate-1], words) {
					logger.Debug("calibration correction", "section", s.Title, "from", s.StartPage, "to", candidate)
					sections[i].StartPage = candidate
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}
	return sections
}

func significantWords(title string) []string {
	var words []string
	for _, w := range strings.Fields(strings.ToLower(title)) {
		if len(w) >= 4 {
			words = append(words, w)
		}
	}
	return words
}

func containsAny(text string, words []string) bool {
	lower := strings.ToLower(text)
	for _, w := range words {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}
