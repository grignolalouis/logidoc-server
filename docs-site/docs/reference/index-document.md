# Index Document

Trigger async indexation. Document must be in `uploaded` or `error` status.

Returns immediately. Poll [Get Document](/reference/get-document) until `ready`.

## Endpoint

```
POST /v1/documents/{id}/index
```

## Response `202 Accepted`

```json
{
  "status": "indexing"
}
```

## Errors

| Code | Reason |
|------|--------|
| `400` | Document is already `ready` or `indexing` |
| `404` | Document not found |

## SDK

::: code-group

```typescript [TypeScript]
await client.indexDocument({ id: doc.id });

// Poll until ready
let d = await client.getDocument({ id: doc.id });
while (d.status === "indexing") {
  await new Promise((r) => setTimeout(r, 3000));
  d = await client.getDocument({ id: doc.id });
}
```

```python [Python]
client.index_document(id=doc.id)

# Poll until ready
import time
d = client.get_document(id=doc.id)
while d.status == "indexing":
    time.sleep(3)
    d = client.get_document(id=doc.id)
```

```go [Go]
c.IndexDocument(ctx, *doc.Id, &logidoc.IndexDocumentRequest{})

// Poll until ready
for {
    d, _ := c.GetDocument(ctx, *doc.Id, &logidoc.GetDocumentRequest{})
    if *d.Status != "indexing" {
        break
    }
    time.Sleep(3 * time.Second)
}
```

```bash [cURL]
curl -X POST http://localhost:7042/v1/documents/5d04eeed-.../index
```

:::
