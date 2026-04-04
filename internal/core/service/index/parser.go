package index

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type PageImage struct {
	Data        []byte
	PageNum     int
	Description string
}

type Pages struct {
	Filename string
	Content  []string    // text per page
	Tables   [][]string  // markdown tables per page (populated if table extraction enabled)
	Images   [][]PageImage // images per page (populated if image extraction enabled)
	PDFPath  string      // temp PDF path (kept alive for image/table extraction)
}

func (p *Pages) Count() int { return len(p.Content) }

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

// ReadEnriched returns text + tables + image descriptions for a page range.
func (p *Pages) ReadEnriched(start, end int) string {
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

		if i < len(p.Tables) {
			for _, tbl := range p.Tables[i] {
				sb.WriteString("\n\n")
				sb.WriteString(tbl)
			}
		}
		if i < len(p.Images) {
			for _, img := range p.Images[i] {
				if img.Description != "" {
					sb.WriteString("\n\n[Image: ")
					sb.WriteString(img.Description)
					sb.WriteString("]")
				}
			}
		}
	}
	return sb.String()
}

// Supported file extensions and their parse strategy.
var (
	pdfExtensions = map[string]bool{".pdf": true}

	// Formats that pandoc converts to markdown.
	pandocExtensions = map[string]bool{
		".docx": true, ".doc": true,
		".pptx": true, ".ppt": true,
		".html": true, ".htm": true,
		".epub": true,
		".xlsx": true, ".xls": true,
		".odt": true, ".ods": true, ".odp": true,
		".rtf": true, ".rst": true, ".org": true,
	}

	textExtensions = map[string]bool{
		".md": true, ".txt": true, ".text": true, ".markdown": true,
	}
)

// SupportedExtension returns true if the file type can be parsed.
func SupportedExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return pdfExtensions[ext] || pandocExtensions[ext] || textExtensions[ext]
}

// ParseFile extracts text from a file. Supports PDF, Office documents, HTML, EPUB, and plaintext.
func ParseFile(filename string, data []byte) (*Pages, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch {
	case pdfExtensions[ext]:
		return parsePDF(filename, data)
	case pandocExtensions[ext]:
		return parsePandoc(filename, data)
	default:
		return &Pages{Filename: filename, Content: []string{string(data)}}, nil
	}
}

// parsePandoc converts any supported format to markdown via pandoc, returns as single page.
func parsePandoc(filename string, data []byte) (*Pages, error) {
	ext := filepath.Ext(filename)
	tmp, err := os.CreateTemp("", "logidoc-*"+ext)
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()

	out, err := exec.Command("pandoc", tmp.Name(), "-t", "markdown", "--wrap=none").Output()
	if err != nil {
		return nil, fmt.Errorf("pandoc convert %s: %w", ext, err)
	}

	text := strings.TrimSpace(string(out))
	if text == "" {
		return nil, fmt.Errorf("pandoc produced empty output for %s", filename)
	}

	return &Pages{Filename: filename, Content: []string{text}}, nil
}

// parsePDF extracts text page by page using pdftotext.
// Empty pages are OCR'd via tesseract if available.
func parsePDF(filename string, data []byte) (*Pages, error) {
	tmp, err := os.CreateTemp("", "logidoc-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	// NOTE: we do NOT defer os.Remove here — PDFPath is kept alive for image extraction.
	// Caller is responsible for cleanup if needed.

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()

	pageCount, err := pdfPageCount(tmp.Name())
	if err != nil {
		os.Remove(tmp.Name())
		return nil, fmt.Errorf("get page count: %w", err)
	}

	content := make([]string, pageCount)
	tables := make([][]string, pageCount)

	for i := 1; i <= pageCount; i++ {
		text, err := pdfExtractPage(tmp.Name(), i)
		if err != nil {
			continue
		}
		text = strings.TrimSpace(text)

		// Heuristic table detection from layout text
		cleanText, pageTables := DetectTables(text)
		content[i-1] = cleanText
		tables[i-1] = pageTables
	}

	return &Pages{
		Filename: filename,
		Content:  content,
		Tables:   tables,
		Images:   make([][]PageImage, pageCount),
		PDFPath:  tmp.Name(),
	}, nil
}

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

// renderPagePNG renders a single PDF page as PNG bytes at 300 DPI.
func renderPagePNG(pdfPath string, page int) ([]byte, error) {
	prefix, err := os.CreateTemp("", "logidoc-render-")
	if err != nil {
		return nil, err
	}
	prefix.Close()
	defer os.Remove(prefix.Name())

	imgPath := prefix.Name() + ".png"
	defer os.Remove(imgPath)

	err = exec.Command("pdftoppm",
		"-f", fmt.Sprintf("%d", page),
		"-l", fmt.Sprintf("%d", page),
		"-png", "-singlefile",
		"-r", "300",
		pdfPath, prefix.Name(),
	).Run()
	if err != nil {
		return nil, fmt.Errorf("pdftoppm page %d: %w", page, err)
	}

	return os.ReadFile(imgPath)
}
