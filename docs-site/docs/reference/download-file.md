# Download File

Returns the original uploaded file.

## Endpoint

```
GET /v1/documents/{id}/file
```

## Response `200 OK`

Returns the raw file with appropriate content type:

| File type | Content-Type |
|-----------|-------------|
| PDF | `application/pdf` |
| Other | `text/plain; charset=utf-8` |

## Errors

| Code | Reason |
|------|--------|
| `404` | Document or file not found |

## SDK

::: code-group

```typescript [TypeScript]
const file = await client.getDocumentFile({ id: doc.id });
// file is a BinaryResponse (ReadableStream)
```

```python [Python]
file = client.get_document_file(id=doc.id)
with open("downloaded.pdf", "wb") as f:
    f.write(file.read())
```

```go [Go]
reader, _ := c.GetDocumentFile(ctx, *doc.Id, &logidoc.GetDocumentFileRequest{})
data, _ := io.ReadAll(reader)
os.WriteFile("downloaded.pdf", data, 0644)
```

```bash [cURL]
curl -o downloaded.pdf http://localhost:7042/v1/documents/5d04eeed-.../file
```

:::
