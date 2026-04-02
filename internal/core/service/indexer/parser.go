package indexer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Pages holds text content extracted per page from a document.
type Pages struct {
	Filename string
	Content  []string // text per page (0-indexed)
}

// Count returns the number of pages.
func (p *Pages) Count() int { return len(p.Content) }

// Read returns the concatenated text for pages from start to end (1-indexed, inclusive).
func (p *Pages) Read(start, end int) string {
	if start < 1 {
		start = 1
	}
	if end > len(p.Content) {
		end = len(p.Content)
	}
	if start > end {
		return ""
	}

	var sb strings.Builder
	for i := start - 1; i < end; i++ {
		if i > start-1 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(p.Content[i])
	}
	return sb.String()
}

// ParseFile extracts text from file content. Supports PDF and plain text.
func ParseFile(filename string, data []byte) (*Pages, error) {
	if strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		return parsePDF(filename, data)
	}
	return &Pages{Filename: filename, Content: []string{string(data)}}, nil
}

// parsePDF uses pdftotext (poppler) to extract text with proper spacing.
// Falls back page-by-page if single extraction fails.
func parsePDF(filename string, data []byte) (*Pages, error) {
	// Write PDF to temp file
	tmp, err := os.CreateTemp("", "logidoc-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()

	// Get page count
	pageCount, err := pdfPageCount(tmp.Name())
	if err != nil {
		return nil, fmt.Errorf("get page count: %w", err)
	}

	// Extract text page by page
	pages := make([]string, pageCount)
	for i := 1; i <= pageCount; i++ {
		text, err := pdfExtractPage(tmp.Name(), i)
		if err != nil {
			continue
		}
		pages[i-1] = strings.TrimSpace(text)
	}

	return &Pages{Filename: filename, Content: pages}, nil
}

// pdfPageCount returns the number of pages using pdfinfo.
func pdfPageCount(path string) (int, error) {
	out, err := exec.Command("pdfinfo", path).Output()
	if err != nil {
		return 0, fmt.Errorf("pdfinfo: %w", err)
	}
	var count int
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "Pages:") {
			fmt.Sscanf(strings.TrimPrefix(line, "Pages:"), "%d", &count)
			break
		}
	}
	if count == 0 {
		return 0, fmt.Errorf("could not determine page count")
	}
	return count, nil
}

// pdfExtractPage extracts text from a single page using pdftotext.
func pdfExtractPage(path string, page int) (string, error) {
	out, err := exec.Command("pdftotext",
		"-f", fmt.Sprintf("%d", page),
		"-l", fmt.Sprintf("%d", page),
		"-layout",
		path, "-",
	).Output()
	if err != nil {
		return "", fmt.Errorf("pdftotext page %d: %w", page, err)
	}
	return string(out), nil
}
