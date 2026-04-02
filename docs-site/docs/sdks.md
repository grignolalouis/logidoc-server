# SDKs

Auto-generated from the [OpenAPI spec](https://github.com/grignolalouis/logidoc-server/blob/main/api/openapi.yaml).

## TypeScript

[GitHub](https://github.com/grignolalouis/logidoc-sdk-ts)

```typescript
import { GrignolalouisApiClient } from "logidoc";

const client = new GrignolalouisApiClient({
  baseUrl: "http://localhost:7042",
});

// Upload
const doc = await client.uploadDocument({
  file: fs.createReadStream("report.pdf"),
});

// Index
await client.indexDocument({ id: doc.id });

// Get TOC
const toc = await client.getDocumentToc({ id: doc.id });

// Get sections
const sections = await client.getDocumentSections({
  id: doc.id,
  ids: "chapter-1,section-2",
});
```

## Go

[GitHub](https://github.com/grignolalouis/logidoc-sdk-go)

```go
import logidoc "github.com/grignolalouis/logidoc-sdk-go"

client := logidoc.NewClient(logidoc.WithBaseURL("http://localhost:7042"))

doc, _ := client.UploadDocument(ctx, file)
client.IndexDocument(ctx, doc.ID)
toc, _ := client.GetDocumentToc(ctx, doc.ID)
sections, _ := client.GetDocumentSections(ctx, doc.ID, "chapter-1,section-2")
```

## Python

[GitHub](https://github.com/grignolalouis/logidoc-sdk-python)

```python
from logidoc import LogidocClient

client = LogidocClient(base_url="http://localhost:7042")

doc = client.upload_document(file=open("report.pdf", "rb"))
client.index_document(id=doc.id)
toc = client.get_document_toc(id=doc.id)
sections = client.get_document_sections(id=doc.id, ids="chapter-1,section-2")
```
