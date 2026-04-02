# MCP Tools

logidoc exposes three MCP tools for AI agents.

## logidoc_list_documents

List all indexed documents.

**Parameters**: none

**Returns**: Markdown-formatted list of documents with IDs, names, and status.

## logidoc_get_toc

Get the table of contents of a document.

**Parameters**:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `doc_id` | string | yes | Document ID |

**Returns**: Indented text tree of sections with IDs, titles, summaries, and page ranges.

```
- [chapter-1] Introduction
  Overview of the project goals and architecture.
  - [section-1-1] Context
    Background and motivation.
  - [section-1-2] Vision
    Long-term goals and design principles.
- [chapter-2] Technical Specifications
  Detailed technical implementation.
```

## logidoc_get_sections

Get the full text of specific sections.

**Parameters**:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `doc_id` | string | yes | Document ID |
| `node_ids` | string | yes | Comma-separated node IDs |

**Returns**: JSON array of sections with full text content.
