# MCP Tools

Three tools exposed via the MCP server at `http://localhost:7043/mcp`.

## logidoc_list_documents

List all indexed documents.

**Parameters**: none

**Returns**: Markdown list of documents with IDs, names, and status.

```
- **report.pdf** [5d04eeed-076d-...] (ready)
  Technical report on the Ora platform.
```

## logidoc_get_toc

Get the table of contents of a document.

**Parameters**:

| Name | Type | Required |
|------|------|----------|
| `doc_id` | string | yes |

**Returns**: Indented text tree.

```
- [chapter-1-introduction] Introduction
  Overview of the project goals and architecture.
  - [section-1-1-context] Context
    Background and motivation for the project.
  - [section-1-2-vision] Vision
    Long-term goals and design principles.
```

## logidoc_get_sections

Get the full text of specific sections.

**Parameters**:

| Name | Type | Required |
|------|------|----------|
| `doc_id` | string | yes |
| `node_ids` | string | yes |

`node_ids` is comma-separated: `"chapter-1,section-1-1"`.

**Returns**: JSON array of nodes with full text content.
