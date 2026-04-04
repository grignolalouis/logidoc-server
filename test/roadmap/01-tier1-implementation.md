# Tier 1 — High Impact

## 1.1 Multi-format support

### Goal
Support docx, pptx, html, epub, xlsx in addition to PDF and plaintext.

### Approach
Use `pandoc` as a universal converter. Everything becomes plaintext or markdown before entering the indexer.

### Architecture

New port in `internal/core/port/document.go`:

```go
type FileConverter interface {
    Convert(ctx context.Context, filename string, data []byte) ([]byte, string, error)
    // Returns: converted data, new filename (e.g. "report.md"), error
}
```

New adapter in `internal/adapter/secondary/converter/pandoc.go`:

```go
package converter

type PandocConverter struct{}

func (c *PandocConverter) Convert(ctx context.Context, filename string, data []byte) ([]byte, string, error) {
    ext := filepath.Ext(filename)
    
    // PDF and text pass through
    if ext == ".pdf" || ext == ".md" || ext == ".txt" {
        return data, filename, nil
    }
    
    // Everything else: pandoc → markdown
    // Write to temp file, run: pandoc input.docx -t markdown -o output.md
    // Return markdown bytes with .md extension
}
```

### Integration

In `indexer/parser.go`, call the converter before parsing:

```go
func ParseFile(filename string, data []byte, converter port.FileConverter) (*Pages, error) {
    data, filename, err := converter.Convert(ctx, filename, data)
    if err != nil {
        return nil, err
    }
    // existing logic: PDF → pdftotext, text → single page
}
```

### Docker

Add to Dockerfile:

```dockerfile
RUN apk add --no-cache poppler-utils pandoc
```

### Supported formats after implementation

| Extension | Converter | Output |
|-----------|-----------|--------|
| .pdf | pdftotext (page by page) | text per page |
| .md, .txt | passthrough | single page |
| .docx | pandoc → markdown | single page |
| .pptx | pandoc → markdown | single page (slides as sections) |
| .html | pandoc → markdown | single page |
| .epub | pandoc → markdown | single page |
| .xlsx | pandoc → markdown | single page (tables as markdown) |

### SOLID compliance
- **S**: FileConverter has one responsibility — format conversion
- **O**: New formats = new converter impl, no changes to existing code
- **D**: Indexer depends on FileConverter interface, not pandoc directly

---

## 1.2 Large node subdivision

### Goal
If a section spans > 20 pages, recursively subdivide it into sub-sections.

### Approach
Post-processing step in Go after tree building. For oversized nodes, send their page range to the LLM to identify sub-structure.

### Architecture

New file `internal/core/service/indexer/subdivide.go`:

```go
package indexer

const maxPagesPerNode = 20

func SubdivideLargeNodes(ctx context.Context, nodes []treeNode, pages *Pages, llm model.Model) []treeNode {
    var result []treeNode
    for _, n := range nodes {
        if n.EndPage - n.StartPage + 1 > maxPagesPerNode && len(n.Children) == 0 {
            // Send this page range to LLM for sub-structure detection
            children := detectSubStructure(ctx, llm, pages, n.StartPage, n.EndPage)
            n.Children = children
        }
        // Recurse into children
        n.Children = SubdivideLargeNodes(ctx, n.Children, pages, llm)
        result = append(result, n)
    }
    return result
}

func detectSubStructure(ctx context.Context, llm model.Model, pages *Pages, start, end int) []treeNode {
    // Reuse the chunking logic: send pages start-end to LLM
    // Ask for sub-sections within this range
    // Returns flat sections → build sub-tree
}
```

### Integration

In `indexer/indexer.go`, after `BuildTree` and before `FillText`:

```go
tree := BuildTree(sections, pages.Count())
tree = SubdivideLargeNodes(ctx, tree, pages, s.llm)  // new step
nodes := FillText(tree, pages)
```

### Config

Add to `IndexerConfig`:

```go
MaxPagesPerNode int  // default 20, configurable via INDEXER_MAX_PAGES_PER_NODE
```

---

## 1.3 OCR fallback

### Goal
Detect scanned PDFs (empty text) and fall back to OCR.

### Approach
After pdftotext, if a page has no text (or very little), render it as image and OCR it.

### Architecture

New port `internal/core/port/document.go`:

```go
type OCRExtractor interface {
    ExtractText(ctx context.Context, imageData []byte) (string, error)
}
```

Two implementations possible:

**Option A: Tesseract** — `internal/adapter/secondary/ocr/tesseract.go`

```go
// Uses tesseract CLI: tesseract input.png output -l eng
// Available in trpc-agent-go: knowledge/ocr/tesseract
```

**Option B: VLM** — `internal/adapter/secondary/ocr/vlm.go`

```go
// Render page as image, send to vision model
// Uses trpc-agent-go: model.Message.AddImageData()
// More accurate for complex layouts, but costs tokens
```

### Integration in parser

```go
func parsePDF(filename string, data []byte, ocr port.OCRExtractor) (*Pages, error) {
    // ... existing pdftotext logic ...
    
    for i, text := range pages {
        if len(strings.TrimSpace(text)) < 50 {
            // Page likely scanned — try OCR
            img := renderPageAsImage(tmpPath, i+1)  // pdftoppm
            if ocrText, err := ocr.ExtractText(ctx, img); err == nil {
                pages[i] = ocrText
            }
        }
    }
}
```

### Docker

```dockerfile
RUN apk add --no-cache poppler-utils tesseract-ocr tesseract-ocr-data-eng
```

### Fallback chain

```
pdftotext → text found? → use it
                        → empty? → pdftoppm (render as image) → tesseract OCR
                                                               → still empty? → VLM vision (if configured)
```

---

## 1.4 Cross-document search

### Goal
New MCP tool `logidoc_search(query)` that searches across all document TOCs.

### Approach
Keyword matching on titles and summaries. No LLM, no vectors — pure Go string matching.

### Architecture

New method on RetrievalService port:

```go
// internal/core/port/retrieval.go
type RetrievalService interface {
    GetTOC(ctx context.Context, docID string) ([]domain.Node, error)
    GetFullTree(ctx context.Context, docID string) ([]domain.Node, error)
    GetSections(ctx context.Context, docID string, nodeIDs []string) ([]domain.Node, error)
    Search(ctx context.Context, query string) ([]domain.SearchHit, error)  // new
}
```

New domain type:

```go
// internal/core/domain/search.go
type SearchHit struct {
    DocID     string
    DocName   string
    NodeID    string
    NodeTitle string
    Summary   string
    StartPage int
    EndPage   int
    Score     float64
}
```

Implementation in `internal/core/service/retrieval_service.go`:

```go
func (s *RetrievalService) Search(ctx context.Context, query string) ([]domain.SearchHit, error) {
    docs, _ := s.docRepo.FindAll(ctx)  // needs docRepo added to service
    
    queryWords := tokenize(query)
    var hits []domain.SearchHit
    
    for _, doc := range docs {
        if doc.Status != domain.StatusReady { continue }
        
        index, err := s.indexRepo.FindByDocID(ctx, doc.ID)
        if err != nil { continue }
        
        walkNodes(index.Tree, func(node domain.Node) {
            score := matchScore(queryWords, node.Title, node.Summary)
            if score > 0 {
                hits = append(hits, domain.SearchHit{
                    DocID: doc.ID, DocName: doc.Name,
                    NodeID: node.ID, NodeTitle: node.Title,
                    Summary: node.Summary, Score: score,
                    StartPage: node.StartPage, EndPage: node.EndPage,
                })
            }
        })
    }
    
    sort.Slice(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
    if len(hits) > 10 { hits = hits[:10] }
    return hits, nil
}

func matchScore(queryWords []string, title, summary string) float64 {
    target := strings.ToLower(title + " " + summary)
    var score float64
    for _, w := range queryWords {
        if strings.Contains(target, w) { score++ }
    }
    return score / float64(len(queryWords))
}
```

### MCP tool

In `internal/adapter/primary/mcp/tools.go`:

```go
mcplib.NewTool("logidoc_search",
    mcplib.WithDescription("Search across all documents. Returns matching sections with doc ID, title, summary, and page range."),
    mcplib.WithString("query", mcplib.Required(), mcplib.Description("Search query")),
)
```

### HTTP endpoint

```
GET /v1/search?q=SSE+streaming
```
