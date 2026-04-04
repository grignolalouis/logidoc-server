# logidoc_search

Search across all indexed documents. Returns the most relevant sections.

## Parameters

| Name | Type | Required |
|------|------|----------|
| `query` | string | yes |

## Returns

Ranked list of matching sections with document name, section title, summary, and page range.

```
- **report.pdf** > SSE Streaming (p.18-19)
  Doc: 5d04eeed-... | Node: sse-streaming
  Server-sent events streaming overview, flow, and implementation.
- **go-book.pdf** > Goroutines (p.217-218)
  Doc: d41ad337-... | Node: goroutines
  Lightweight concurrent threads managed by the Go runtime.
```

## When to use

When you don't know which document contains the information. Searches across all TOCs using keyword matching on titles and summaries. No LLM cost.

After finding relevant sections, call `logidoc_get_sections` to retrieve the full text.
