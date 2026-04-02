# API Endpoints

Base URL: `http://localhost:7042`

## Documents

### Upload a document

```
POST /v1/documents
Content-Type: multipart/form-data
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | binary | yes | PDF, Markdown, or text file |

**Response** `201`

```json
{
  "id": "5d04eeed-076d-44bf-a924-c9df4060ceba",
  "name": "report.pdf",
  "status": "uploaded",
  "created_at": "2026-04-02T10:00:00Z"
}
```

### List documents

```
GET /v1/documents
```

Returns documents with status `ready`.

### Get document

```
GET /v1/documents/{id}
```

### Delete document

```
DELETE /v1/documents/{id}
```

### Download file

```
GET /v1/documents/{id}/file
```

Returns the original uploaded file.

## Indexation

### Trigger indexation

```
POST /v1/documents/{id}/index
```

Starts async indexation. Document must be in `uploaded` or `error` status.

**Response** `202`

```json
{ "status": "indexing" }
```

Poll `GET /v1/documents/{id}` until status is `ready`.

## Retrieval

### Get table of contents

```
GET /v1/documents/{id}/toc
```

Returns the hierarchical section tree with titles, summaries, and page ranges. No full text.

**Response**

```json
{
  "toc": [
    {
      "id": "chapter-1",
      "title": "Introduction",
      "summary": "Overview of the project goals and architecture.",
      "start_page": 1,
      "end_page": 5,
      "children": [
        {
          "id": "section-1-1",
          "title": "Context",
          "summary": "Background and motivation.",
          "start_page": 1,
          "end_page": 2,
          "children": []
        }
      ]
    }
  ]
}
```

### Get sections

```
GET /v1/documents/{id}/sections?ids=chapter-1,section-1-1
```

Returns the full text of the specified sections.

## Health

### Health check

```
GET /health
```

```json
{
  "status": "ok",
  "version": "1.0.0",
  "mongo": "ok"
}
```

### Version

```
GET /version
```
