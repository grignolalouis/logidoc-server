# List Documents

Returns all documents with status `ready`.

## Endpoint

```
GET /v1/documents
```

## Response `200 OK`

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

## SDK

::: code-group

```typescript [TypeScript]
const docs = await client.listDocuments();
for (const doc of docs.documents) {
  console.log(`${doc.name} [${doc.status}]`);
}
```

```python [Python]
docs = client.list_documents()
for doc in docs.documents:
    print(f"{doc.name} [{doc.status}]")
```

```go [Go]
docs, _ := c.ListDocuments(ctx)
for _, doc := range docs.Documents {
    fmt.Printf("%s [%s]\n", *doc.Name, *doc.Status)
}
```

```bash [cURL]
curl http://localhost:7042/v1/documents
```

:::
