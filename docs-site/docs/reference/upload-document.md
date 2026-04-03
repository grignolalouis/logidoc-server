# Upload Document

Upload a file. Does not trigger indexation — call [Index Document](/reference/index-document) after.

## Endpoint

```
POST /v1/documents
Content-Type: multipart/form-data
```

| Field | Type | Required |
|-------|------|----------|
| `file` | binary | yes |

## Response `201 Created`

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

## SDK

::: code-group

```typescript [TypeScript]
import fs from "fs";

const doc = await client.uploadDocument({
  file: fs.createReadStream("report.pdf"),
});
// doc.id → "5d04eeed-076d-..."
// doc.status → "uploaded"
```

```python [Python]
doc = client.upload_document(file=open("report.pdf", "rb"))
# doc.id → "5d04eeed-076d-..."
# doc.status → "uploaded"
```

```go [Go]
f, _ := os.Open("report.pdf")
defer f.Close()

doc, _ := c.UploadDocument(ctx, f)
// *doc.Id → "5d04eeed-076d-..."
// *doc.Status → "uploaded"
```

```bash [cURL]
curl -X POST http://localhost:7042/v1/documents \
  -F "file=@report.pdf"
```

:::
