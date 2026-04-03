# Delete Document

Removes the document, its stored file, and its index.

## Endpoint

```
DELETE /v1/documents/{id}
```

## Response `204 No Content`

## Errors

| Code | Reason |
|------|--------|
| `404` | Document not found |

## SDK

::: code-group

```typescript [TypeScript]
await client.deleteDocument({ id: doc.id });
```

```python [Python]
client.delete_document(id=doc.id)
```

```go [Go]
c.DeleteDocument(ctx, *doc.Id, &logidoc.DeleteDocumentRequest{})
```

```bash [cURL]
curl -X DELETE http://localhost:7042/v1/documents/5d04eeed-...
```

:::
