<p align="center">
  <img src=".github/banner.png" alt="logidoc" width="100%">
</p>

<p align="right">
  Open source, self-hosted alternative to PageIndex.<br>
  Document indexing for AI agents. No vectors, no embeddings.<br>
  Upload a PDF, get a searchable table of contents.<br>
  <a href="https://grignolalouis.github.io/logidoc-server/">Read the docs to get started →</a>
</p>

## MCP

Add to your project's `.mcp.json` and restart your client:

```json
{
  "mcpServers": {
    "logidoc": {
      "type": "http",
      "url": "http://localhost:7043/mcp"
    }
  }
}
```

Works with Claude Code, Cursor, Windsurf, and any MCP-compatible client.

**Available tools:**

| Tool | Description |
|------|-------------|
| `logidoc_list_documents` | List all indexed documents |
| `logidoc_get_toc` | Get table of contents (titles, summaries, pages) |
| `logidoc_get_sections` | Get full text of specific sections |

## SDKs

**TypeScript**

```typescript
import { GrignolalouisApiClient } from "logidoc";

const client = new GrignolalouisApiClient({ baseUrl: "http://localhost:7042" });

const doc = await client.uploadDocument({ file: fs.createReadStream("report.pdf") });
await client.indexDocument({ id: doc.id });
const toc = await client.getDocumentToc({ id: doc.id });
const sections = await client.getDocumentSections({ id: doc.id, ids: "chapter-1" });
```

**Python**

```python
from logidoc import LogidocClient

client = LogidocClient(base_url="http://localhost:7042")

doc = client.upload_document(file=open("report.pdf", "rb"))
client.index_document(id=doc.id)
toc = client.get_document_toc(id=doc.id)
sections = client.get_document_sections(id=doc.id, ids="chapter-1")
```

**Go**

```go
client := logidoc.NewClient(logidoc.WithBaseURL("http://localhost:7042"))

doc, _ := client.UploadDocument(ctx, file)
client.IndexDocument(ctx, doc.ID)
toc, _ := client.GetDocumentToc(ctx, doc.ID)
sections, _ := client.GetDocumentSections(ctx, doc.ID, "chapter-1")
```

**Repositories:** [Go](https://github.com/grignolalouis/logidoc-sdk-go) · [TypeScript](https://github.com/grignolalouis/logidoc-sdk-ts) · [Python](https://github.com/grignolalouis/logidoc-sdk-python)

## License

MIT
