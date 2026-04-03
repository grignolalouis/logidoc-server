# Get Sections

Retrieve the full text of specific sections by node IDs from the TOC.

## Endpoint

```
GET /v1/documents/{id}/sections?ids={node_ids}
```

| Parameter | In | Required | Description |
|-----------|-----|----------|-------------|
| `id` | path | yes | Document ID |
| `ids` | query | yes | Comma-separated node IDs |

## Response `200 OK`

```json
{
  "sections": [
    {
      "id": "section-1-1-context",
      "title": "Context",
      "summary": "Background and motivation for the project.",
      "text": "Ora is a platform designed for interacting with conversational AI agents. The project is composed of two distinct components...",
      "start_page": 1,
      "end_page": 3,
      "children": []
    }
  ]
}
```

::: tip
Only the requested nodes are returned. Children are not recursively populated.
:::

## Errors

| Code | Reason |
|------|--------|
| `400` | Missing `ids` parameter |
| `404` | Document or index not found |

## SDK

::: code-group

```typescript [TypeScript]
const sections = await client.getDocumentSections({
  id: doc.id,
  ids: "chapter-1-introduction,section-1-1-context",
});
for (const s of sections.sections) {
  console.log(`## ${s.title} (p.${s.startPage})`);
  console.log(s.text);
}
```

```python [Python]
sections = client.get_document_sections(
    id=doc.id,
    ids="chapter-1-introduction,section-1-1-context",
)
for s in sections.sections:
    print(f"## {s.title} (p.{s.start_page})")
    print(s.text)
```

```go [Go]
sections, _ := c.GetDocumentSections(ctx, *doc.Id, &logidoc.GetDocumentSectionsRequest{
    IDs: "chapter-1-introduction,section-1-1-context",
})
for _, s := range sections.Sections {
    fmt.Printf("## %s (p.%d)\n", *s.Title, *s.StartPage)
    fmt.Println(*s.Text)
}
```

```bash [cURL]
curl "http://localhost:7042/v1/documents/5d04eeed-.../sections?ids=chapter-1-introduction,section-1-1-context"
```

:::
