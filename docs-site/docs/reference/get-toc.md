# Get TOC

Returns the hierarchical table of contents. Each node has a title, summary, and page range. No full text.

## Endpoint

```
GET /v1/documents/{id}/toc
```

## Response `200 OK`

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
        }
      ]
    }
  ]
}
```

## Node schema

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique kebab-case identifier |
| `title` | string | Section heading |
| `summary` | string | One sentence, 15–25 words |
| `text` | string | Always empty in TOC response |
| `start_page` | int | First page (1-indexed) |
| `end_page` | int | Last page (1-indexed) |
| `children` | Node[] | Nested sub-sections |

## Errors

| Code | Reason |
|------|--------|
| `404` | Document or index not found |

## SDK

::: code-group

```typescript [TypeScript]
const toc = await client.getDocumentToc({ id: doc.id });
for (const section of toc.toc) {
  console.log(`${section.title} (p.${section.startPage}-${section.endPage})`);
}
```

```python [Python]
toc = client.get_document_toc(id=doc.id)
for section in toc.toc:
    print(f"{section.title} (p.{section.start_page}-{section.end_page})")
```

```go [Go]
toc, _ := c.GetDocumentToc(ctx, *doc.Id, &logidoc.GetDocumentTocRequest{})
for _, s := range toc.Toc {
    fmt.Printf("%s (p.%d-%d)\n", *s.Title, *s.StartPage, *s.EndPage)
}
```

```bash [cURL]
curl http://localhost:7042/v1/documents/5d04eeed-.../toc
```

:::
