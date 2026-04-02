# Get Document

Get details and status of a document.

## Endpoint

```
GET /v1/documents/{id}
```

## Response `200 OK`

```json
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
```

## Errors

| Code | Reason |
|------|--------|
| `404` | Document not found |

## SDK

::: code-group

```typescript [TypeScript]
const doc = await client.getDocument({ id: "5d04eeed-..." });
console.log(doc.status); // "ready"
```

```python [Python]
doc = client.get_document(id="5d04eeed-...")
print(doc.status)  # "ready"
```

```go [Go]
doc, _ := client.GetDocument(ctx, "5d04eeed-...")
fmt.Println(doc.Status) // "ready"
```

```bash [cURL]
curl http://localhost:7042/v1/documents/5d04eeed-...
```

:::

## Status lifecycle

```
uploaded → indexing → ready
                   → error
```

Use this endpoint to poll indexation progress.
