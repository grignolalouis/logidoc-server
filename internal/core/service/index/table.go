package index

import (
	"context"
	"log/slog"
	"strings"

	"trpc.group/trpc-go/trpc-agent-go/model"
)

// DetectTables scans pdftotext -layout output for table-like regions and converts them to markdown.
// Returns the cleaned text (tables removed) and extracted markdown tables.
func DetectTables(pageText string) (string, []string) {
	lines := strings.Split(pageText, "\n")
	if len(lines) < 3 {
		return pageText, nil
	}

	var tables []string
	var cleanLines []string
	var tableLines []string
	inTable := false

	for _, line := range lines {
		if isTabularLine(line) {
			if !inTable {
				inTable = true
				tableLines = nil
			}
			tableLines = append(tableLines, line)
		} else {
			if inTable && len(tableLines) >= 2 {
				if md := convertToMarkdown(tableLines); md != "" {
					tables = append(tables, md)
				}
			}
			inTable = false
			tableLines = nil
			cleanLines = append(cleanLines, line)
		}
	}

	// Flush last table
	if inTable && len(tableLines) >= 2 {
		if md := convertToMarkdown(tableLines); md != "" {
			tables = append(tables, md)
		}
	}

	return strings.Join(cleanLines, "\n"), tables
}

// isTabularLine detects lines with multiple whitespace-separated columns.
// A line is "tabular" if it has 2+ gaps of 3+ spaces between non-empty segments.
func isTabularLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 5 {
		return false
	}

	gaps := 0
	inGap := false
	gapLen := 0

	for _, ch := range trimmed {
		if ch == ' ' {
			gapLen++
			if gapLen >= 3 && !inGap {
				gaps++
				inGap = true
			}
		} else {
			gapLen = 0
			inGap = false
		}
	}

	return gaps >= 2
}

// convertToMarkdown turns aligned text lines into a markdown table.
func convertToMarkdown(lines []string) string {
	if len(lines) < 2 {
		return ""
	}

	// Find column boundaries by detecting consistent whitespace positions
	columns := detectColumns(lines)
	if len(columns) < 2 {
		return ""
	}

	var rows [][]string
	for _, line := range lines {
		row := splitByColumns(line, columns)
		rows = append(rows, row)
	}

	if len(rows) < 2 {
		return ""
	}

	// Build markdown table
	var sb strings.Builder
	numCols := len(columns)

	// Header
	sb.WriteString("|")
	for _, cell := range padRow(rows[0], numCols) {
		sb.WriteString(" ")
		sb.WriteString(strings.TrimSpace(cell))
		sb.WriteString(" |")
	}
	sb.WriteString("\n")

	// Separator
	sb.WriteString("|")
	for range numCols {
		sb.WriteString("---|")
	}
	sb.WriteString("\n")

	// Data rows
	for _, row := range rows[1:] {
		sb.WriteString("|")
		for _, cell := range padRow(row, numCols) {
			sb.WriteString(" ")
			sb.WriteString(strings.TrimSpace(cell))
			sb.WriteString(" |")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func detectColumns(lines []string) []int {
	if len(lines) == 0 {
		return nil
	}

	maxLen := 0
	for _, l := range lines {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}

	// Count how many lines have a space at each position
	spaceCounts := make([]int, maxLen)
	for _, line := range lines {
		for i, ch := range line {
			if ch == ' ' {
				spaceCounts[i]++
			}
		}
	}

	// Column boundaries = positions where most lines have spaces
	threshold := len(lines) * 2 / 3
	var boundaries []int
	inGap := false

	for i, count := range spaceCounts {
		if count >= threshold {
			if !inGap {
				boundaries = append(boundaries, i)
				inGap = true
			}
		} else {
			inGap = false
		}
	}

	// Add start position
	result := []int{0}
	result = append(result, boundaries...)
	return result
}

func splitByColumns(line string, columns []int) []string {
	cells := make([]string, len(columns))
	for i, start := range columns {
		end := len(line)
		if i+1 < len(columns) {
			end = columns[i+1]
		}
		if start < len(line) {
			if end > len(line) {
				end = len(line)
			}
			cells[i] = line[start:end]
		}
	}
	return cells
}

func padRow(row []string, n int) []string {
	for len(row) < n {
		row = append(row, "")
	}
	return row[:n]
}

func enrichTablesVLM(ctx context.Context, llm model.Model, pages *Pages, metrics *Metrics, logger *slog.Logger) {
	for i := range pages.Content {
		if len(pages.Tables[i]) > 0 {
			continue
		}
		if !looksTabular(pages.Content[i]) {
			continue
		}
		vlmTables, usage, err := ExtractTablesVLM(ctx, llm, pages.PDFPath, i+1)
		if err != nil {
			logger.Warn("vlm table extraction failed", "page", i+1, "error", err)
			continue
		}
		metrics.AddVision(usage)
		if len(vlmTables) > 0 {
			pages.Tables[i] = append(pages.Tables[i], vlmTables...)
			logger.Debug("vlm tables extracted", "page", i+1, "count", len(vlmTables))
		}
	}
}

func looksTabular(text string) bool {
	lines := strings.Split(text, "\n")
	if len(lines) < 4 {
		return false
	}
	// Count lines that have multiple whitespace gaps (column-like alignment)
	alignedLines := 0
	for _, l := range lines {
		gaps := 0
		inGap := false
		gapLen := 0
		for _, ch := range l {
			if ch == ' ' {
				gapLen++
				if gapLen >= 2 && !inGap {
					gaps++
					inGap = true
				}
			} else {
				gapLen = 0
				inGap = false
			}
		}
		if gaps >= 1 && len(strings.TrimSpace(l)) > 5 {
			alignedLines++
		}
	}
	// At least 60% of lines should look column-aligned
	return alignedLines > len(lines)*3/5
}

const extractTablePrompt = "Extract ALL tables from this page as markdown tables. If there are no tables, respond with 'NO_TABLES'. Output only the markdown tables, nothing else."

// ExtractTablesVLM renders a page as image and asks the VLM to extract tables as markdown.
func ExtractTablesVLM(ctx context.Context, llm model.Model, pdfPath string, page int) ([]string, *model.Usage, error) {
	imgData, err := RenderPageAsImage(pdfPath, page)
	if err != nil {
		return nil, nil, err
	}

	content, usage, err := CallVision(ctx, llm, imgData, extractTablePrompt)
	if err != nil {
		return nil, nil, err
	}

	if strings.Contains(content, "NO_TABLES") || len(content) < 10 {
		return nil, usage, nil
	}

	return []string{content}, usage, nil
}
