# logidoc_get_toc

Get the table of contents of a document.

## Parameters

| Name | Type | Required |
|------|------|----------|
| `doc_id` | string | yes |

## Returns

Indented text tree with section IDs, titles, summaries, and page ranges.

```
- [chapter-1-introduction] Introduction
  Overview of the project goals and architecture.
  - [section-1-1-context] Context (p.1-3)
    Background and motivation for the project.
  - [section-1-2-vision] Vision (p.4-5)
    Long-term goals and design principles.
- [chapter-2-technical] Technical Specifications (p.15-32)
  Architecture, API, testing, and deployment details.
```

## When to use

After listing documents, read the TOC to understand a document's structure. Use the summaries to decide which sections are relevant to the user's question, then call `logidoc_get_sections` with the node IDs.

The TOC is lightweight (~2k tokens for a 400-page book). The full text is NOT included — only titles and summaries.
