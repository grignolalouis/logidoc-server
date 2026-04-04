# Tier 2 — Medium Impact

## 2.1 Table extraction

### Goal
Extract tables from PDFs as structured markdown that agents can read and query.

### Approach
Two-tier extraction: layout-based first, VLM fallback for complex tables.

### Architecture

New port:

```go
// internal/core/port/document.go
type TableExtractor interface {
    ExtractTables(ctx context.Context, pdfPath string, page int) ([]string, error)
    // Returns markdown tables found on this page
}
```

**Implementation A: Layout-based** — `internal/adapter/secondary/table/layout.go`

```go
// Uses pdftotext -layout and post-processes the output
// Detects aligned columns by character positions
// Converts to markdown table format
// Fast, free, works for simple tables
```

**Implementation B: VLM-based** — `internal/adapter/secondary/table/vlm.go`

```go
// Render page as image via pdftoppm
// Send to vision model: "Extract all tables from this page as markdown"
// Uses trpc-agent-go: model.Message.AddImageData()
// Accurate for complex tables, costs tokens
```

### Integration in parser

Each page's content becomes: `text + "\n\n" + tables_markdown`

```go
// internal/core/service/indexer/parser.go
type Pages struct {
    Filename string
    Content  []string // text per page
    Tables   []string // markdown tables per page (parallel array)
}

func (p *Pages) Read(start, end int) string {
    var sb strings.Builder
    for i := start - 1; i < end; i++ {
        sb.WriteString(p.Content[i])
        if p.Tables[i] != "" {
            sb.WriteString("\n\n")
            sb.WriteString(p.Tables[i])
        }
    }
    return sb.String()
}
```

### Fallback chain

```
pdftotext -layout → detect table patterns → markdown
    ↓ (if complex/misaligned)
pdftoppm → image → VLM "extract tables as markdown"
```

---

## 2.2 Image extraction + VLM description

### Goal
Extract images/diagrams from PDFs, describe them via a vision model, and include descriptions in the indexed content.

### Approach
Extract images with `pdfimages`, send to VLM for description, append to page content.

### Architecture

New port:

```go
// internal/core/port/document.go
type ImageDescriber interface {
    Describe(ctx context.Context, imageData []byte) (string, error)
}
```

Implementation using trpc-agent-go vision:

```go
// internal/adapter/secondary/vision/describer.go
package vision

import "trpc.group/trpc-go/trpc-agent-go/model"

type Describer struct {
    llm model.Model
}

func (d *Describer) Describe(ctx context.Context, imageData []byte) (string, error) {
    msg := model.NewUserMessage("Describe this image concisely. Focus on data, labels, and structure.")
    msg.AddImageData(imageData, "auto", "png")
    
    req := &model.Request{
        Messages: []model.Message{msg},
        GenerationConfig: model.GenerationConfig{Stream: false},
    }
    
    respChan, err := d.llm.GenerateContent(ctx, req)
    // collect response...
}
```

### Integration in parser

```go
func parsePDF(filename string, data []byte, imgDescriber port.ImageDescriber) (*Pages, error) {
    // ... existing text extraction ...
    
    for i := range pages {
        images := extractPageImages(tmpPath, i+1)  // pdfimages -png
        for _, img := range images {
            desc, err := imgDescriber.Describe(ctx, img)
            if err == nil {
                pages[i] += "\n\n[Image: " + desc + "]"
            }
        }
    }
}
```

### Config

```env
ENABLE_IMAGE_DESCRIPTION=true   # disabled by default (costs tokens)
```

### Docker

```dockerfile
RUN apk add --no-cache poppler-utils  # pdfimages already included
```

---

## 2.3 Better summaries

### Goal
Generate summaries from actual section content, not just titles.

### Approach
Post-indexation step: for each leaf node, send its text to the LLM for a 15-25 word summary.

### Architecture

New file `internal/core/service/indexer/summarize.go`:

```go
package indexer

func SummarizeNodes(ctx context.Context, nodes []treeNode, pages *Pages, llm model.Model) []treeNode {
    for i, n := range nodes {
        if n.Summary == "" || isTitleOnly(n.Summary, n.Title) {
            text := pages.Read(n.StartPage, n.EndPage)
            if len(text) > 2000 {
                text = text[:2000] // first 2000 chars enough for summary
            }
            summary := generateSummary(ctx, llm, n.Title, text)
            nodes[i].Summary = summary
        }
        nodes[i].Children = SummarizeNodes(ctx, n.Children, pages, llm)
    }
    return nodes
}

func generateSummary(ctx context.Context, llm model.Model, title, text string) string {
    prompt := fmt.Sprintf(
        "Write a one-sentence summary (15-25 words) of this section titled %q:\n\n%s",
        title, text,
    )
    // Single LLM call, ~500 tokens input, ~30 tokens output
}
```

### Integration

In `indexer/indexer.go`, after subdivision, before FillText:

```go
tree := BuildTree(sections, pages.Count())
tree = SubdivideLargeNodes(ctx, tree, pages, s.llm)
tree = SummarizeNodes(ctx, tree, pages, s.llm)  // new
nodes := FillText(tree, pages)
```

### Token cost
~500 tokens per section. 50 sections = ~25k tokens. Acceptable for better retrieval accuracy.

### Config

```env
ENABLE_CONTENT_SUMMARIES=true  # disabled by default
```

---

## 2.4 Page calibration improvement

### Goal
Verify and correct page boundaries by matching section headings in the extracted text.

### Approach
After CalibratePages (offset detection), verify each section's title appears on its assigned page. If not, search nearby pages.

### Architecture

New function in `internal/core/service/indexer/tree.go`:

```go
func VerifyPageBoundaries(sections []FlatSection, pages *Pages) []FlatSection {
    for i, s := range sections {
        pageText := strings.ToLower(pages.Content[s.StartPage-1])
        titleWords := significantWords(s.Title)
        
        if !containsAny(pageText, titleWords) {
            // Title not found on assigned page — search nearby
            for offset := 1; offset <= 3; offset++ {
                for _, delta := range []int{-offset, offset} {
                    candidate := s.StartPage + delta
                    if candidate < 1 || candidate > pages.Count() { continue }
                    if containsAny(strings.ToLower(pages.Content[candidate-1]), titleWords) {
                        sections[i].StartPage = candidate
                        break
                    }
                }
            }
        }
    }
    return sections
}
```

### Integration

In `indexer/indexer.go`, after CalibratePages:

```go
sections = CalibratePages(sections, pages)
sections = VerifyPageBoundaries(sections, pages)  // new
tree := BuildTree(sections, pages.Count())
```

Zero LLM cost — pure Go string matching.
