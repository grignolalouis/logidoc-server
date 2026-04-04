# Tier 3 — Long Term

## 3.1 Streaming indexation status

### Goal
Real-time progress updates during indexation via SSE.

### Approach
The indexer emits events (parsing, TOC detection, chunking, saving) to a channel. An SSE endpoint streams these to the client.

### Architecture

New domain type:

```go
// internal/core/domain/event.go
type IndexEvent struct {
    DocID   string
    Phase   string // "parsing", "toc_detection", "chunking", "saving", "done", "error"
    Message string
    Progress float64 // 0.0 to 1.0
}
```

New port:

```go
// internal/core/port/document.go
type EventBus interface {
    Publish(event domain.IndexEvent)
    Subscribe(docID string) <-chan domain.IndexEvent
    Unsubscribe(docID string, ch <-chan domain.IndexEvent)
}
```

Implementation: in-memory pub/sub with `sync.Map` of channels per doc ID.

```go
// internal/infrastructure/eventbus/eventbus.go
package eventbus

type Bus struct {
    subs sync.Map // map[string][]chan domain.IndexEvent
}
```

### HTTP endpoint

```
GET /v1/documents/{id}/events
Content-Type: text/event-stream

data: {"phase": "parsing", "progress": 0.1, "message": "Extracting text from 32 pages"}
data: {"phase": "toc_detection", "progress": 0.3, "message": "Reading first pages for TOC"}
data: {"phase": "done", "progress": 1.0, "message": "Indexed 51 sections"}
```

### UI integration

Replace HTMX polling with EventSource:

```javascript
const es = new EventSource(`/v1/documents/${docId}/events`);
es.onmessage = (e) => {
    const event = JSON.parse(e.data);
    updateProgress(event.phase, event.progress);
    if (event.phase === "done" || event.phase === "error") {
        es.close();
        location.reload();
    }
};
```

---

## 3.2 Document versioning

### Goal
Track changes between re-indexations of the same document.

### Approach
Store previous index versions in MongoDB. Diff TOCs to show what changed.

### Architecture

Extend Index domain:

```go
type Index struct {
    DocID     string
    Tree      []Node
    Version   int
    CreatedAt time.Time
    PrevTree  []Node // previous version for diff
}
```

New method:

```go
// internal/core/service/retrieval_service.go
func (s *RetrievalService) GetDiff(ctx context.Context, docID string) (*domain.IndexDiff, error) {
    // Compare current tree with previous
    // Return: added sections, removed sections, modified sections
}
```

### Mongo storage

```go
// Keep last N versions in a separate collection "index_versions"
// On re-index: move current to versions, save new as current
```

---

## 3.3 Token cost tracking

### Goal
Track and expose LLM token usage per document.

### Approach
The indexer already collects Metrics (prompt_tokens, completion_tokens, total_tokens). Store them in the document record.

### Architecture

Extend Document domain:

```go
type Document struct {
    // ... existing fields ...
    IndexMetrics *IndexMetrics
}

type IndexMetrics struct {
    Duration         time.Duration
    LLMCalls         int
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
    PagesRead        int
}
```

Store in MongoDB alongside the document. Expose via API:

```
GET /v1/documents/{id} → includes "metrics" field
```

### UI

Show metrics on the document detail page: tokens used, duration, LLM calls, cost estimate.

---

## 3.4 Incremental re-indexing

### Goal
Update specific sections without re-processing the entire document.

### Approach
This is complex. Two strategies:

**Strategy A: Section-level re-index**

```
POST /v1/documents/{id}/reindex?sections=chapter-2,chapter-3
```

Only re-processes the specified sections (their page ranges). Merges updated nodes into existing tree.

**Strategy B: Append-only**

For documents that grow (logs, reports), allow appending new pages:

```
POST /v1/documents/{id}/append
Content-Type: multipart/form-data
file: new-pages.pdf
```

Parse only the new pages, detect structure, merge into existing tree.

### Architecture

Both strategies require:
1. Ability to merge partial trees into existing index
2. Re-calculation of end_page boundaries
3. Re-generation of summaries for affected sections

```go
// internal/core/service/indexer/merge.go
func MergeTrees(existing []treeNode, updated []treeNode, replaceIDs []string) []treeNode {
    // Walk existing tree
    // For each node in replaceIDs, swap with corresponding updated node
    // Recalculate page boundaries
}
```

### Complexity
High. Depends on 3.2 (versioning) for rollback capability. Implement last.

---

## Implementation order

```
3.3 Token cost tracking     ← easy, metrics already collected
3.1 Streaming status        ← medium, needs eventbus
3.2 Document versioning     ← medium, needs mongo schema change
3.4 Incremental re-index    ← hard, depends on 3.2
```
