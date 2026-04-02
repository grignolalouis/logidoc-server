# API Reference

Base URL: `http://localhost:7042`

All `/v1/*` endpoints require `Authorization: Bearer <API_KEY>` when `API_KEY` is configured.

## Upload document

Upload a file. Does not trigger indexation.

```
POST /v1/documents
Content-Type: multipart/form-data
```

**Request**

| Field | Type | Required |
|-------|------|----------|
| `file` | binary | yes |

**Response** `201 Created`

```json
{
  "id": "5d04eeed-076d-44bf-a924-c9df4060ceba",
  "name": "report.pdf",
  "description": "",
  "status": "uploaded",
  "page_count": 0,
  "node_count": 0,
  "error": null,
  "created_at": "2026-04-02T10:00:00Z",
  "indexed_at": null
}
```

---

## List documents

Returns documents with status `ready`.

```
GET /v1/documents
```

**Response** `200 OK`

```json
{
  "documents": [
    {
      "id": "5d04eeed-076d-44bf-a924-c9df4060ceba",
      "name": "report.pdf",
      "description": "Technical report on the Ora platform.",
      "status": "ready",
      "page_count": 0,
      "node_count": 0,
      "error": null,
      "created_at": "2026-04-02T10:00:00Z",
      "indexed_at": "2026-04-02T10:00:45Z"
    }
  ]
}
```

---

## Get document

```
GET /v1/documents/{id}
```

**Response** `200 OK` — same shape as upload response.

**Response** `404 Not Found`

```json
{
  "error": "document not found",
  "message": "document not found"
}
```

---

## Download file

Returns the original uploaded file.

```
GET /v1/documents/{id}/file
```

**Response** `200 OK`

- `Content-Type: application/pdf` for PDFs
- `Content-Type: text/plain` for text/markdown
- `Content-Disposition: inline; filename="report.pdf"`

---

## Trigger indexation

Starts async indexation. Document must be `uploaded` or `error`.

```
POST /v1/documents/{id}/index
```

**Response** `202 Accepted`

```json
{
  "status": "indexing"
}
```

**Response** `400 Bad Request`

```json
{
  "error": "document 5d04eeed is ready, cannot index"
}
```

Poll `GET /v1/documents/{id}` until `status` is `ready` or `error`.

### Status lifecycle

```
uploaded → indexing → ready
                   → error (can retry: POST /index again)
```

---

## Delete document

Removes the document, its stored file, and its index.

```
DELETE /v1/documents/{id}
```

**Response** `204 No Content`

---

## Get table of contents

Returns the hierarchical tree with titles, summaries, and page ranges. No full text.

```
GET /v1/documents/{id}/toc
```

**Response** `200 OK`

```json
{
  "toc": [
    {
      "id": "chapter-1-introduction",
      "title": "Introduction",
      "summary": "Overview of the project goals and architecture.",
      "text": "",
      "start_page": 1,
      "end_page": 14,
      "children": [
        {
          "id": "section-1-1-context",
          "title": "Context",
          "summary": "Background and motivation for the project.",
          "text": "",
          "start_page": 1,
          "end_page": 3,
          "children": []
        },
        {
          "id": "section-1-2-vision",
          "title": "Vision",
          "summary": "Long-term goals and design principles.",
          "text": "",
          "start_page": 4,
          "end_page": 5,
          "children": []
        }
      ]
    },
    {
      "id": "chapter-2-technical",
      "title": "Technical Specifications",
      "summary": "Architecture, API, testing, and deployment details.",
      "text": "",
      "start_page": 15,
      "end_page": 32,
      "children": []
    }
  ]
}
```

### Node schema

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique kebab-case identifier |
| `title` | string | Section heading |
| `summary` | string | One sentence, 15–25 words |
| `text` | string | Empty in TOC responses |
| `start_page` | integer | First page (1-indexed) |
| `end_page` | integer | Last page (1-indexed) |
| `children` | Node[] | Nested sub-sections |

---

## Get sections

Returns full text for specific sections.

```
GET /v1/documents/{id}/sections?ids=chapter-1-introduction,section-1-1-context
```

**Parameters**

| Name | In | Required | Description |
|------|----|----------|-------------|
| `id` | path | yes | Document ID |
| `ids` | query | yes | Comma-separated node IDs |

**Response** `200 OK`

```json
{
  "sections": [
    {
      "id": "chapter-1-introduction",
      "title": "Introduction",
      "summary": "Overview of the project goals and architecture.",
      "text": "Ora is a platform designed for interacting with conversational AI agents. The project is composed of two distinct components...",
      "start_page": 1,
      "end_page": 14,
      "children": []
    },
    {
      "id": "section-1-1-context",
      "title": "Context",
      "summary": "Background and motivation for the project.",
      "text": "The mobile application serves as the primary interface for end users to communicate with various AI agents...",
      "start_page": 1,
      "end_page": 3,
      "children": []
    }
  ]
}
```

::: tip
Sections returned by this endpoint include `text` with the full raw content. Children are not recursively populated — only the requested nodes are returned.
:::

---

## Health check

```
GET /health
```

**Response** `200 OK`

```json
{
  "status": "ok",
  "version": "1.0.0",
  "mongo": "ok"
}
```

| status | Meaning |
|--------|---------|
| `ok` | All systems operational |
| `degraded` | MongoDB unreachable |

---

## Version

```
GET /version
```

**Response** `200 OK`

```json
{
  "version": "1.0.0"
}
```

---

## Error format

All errors follow the same shape:

```json
{
  "error": "short error description",
  "message": "detailed error message"
}
```

| HTTP Code | Meaning |
|-----------|---------|
| `400` | Bad request (invalid input, wrong document status) |
| `401` | Missing or invalid API key |
| `404` | Document or index not found |
| `409` | Document not yet indexed |
| `429` | Rate limit exceeded |
| `500` | Internal server error |
