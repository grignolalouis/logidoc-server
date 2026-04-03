<p align="center">
  <img src=".github/banner.png" alt="logidoc" width="100%">
</p>

Open source, self-hosted alternative to PageIndex.
Document indexing for AI agents. No vectors, no embeddings.
Upload a PDF, get a searchable table of contents.

[Read the docs to get started →](https://grignolalouis.github.io/logidoc-server/)

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

[Go](https://github.com/grignolalouis/logidoc-sdk-go) · [TypeScript](https://github.com/grignolalouis/logidoc-sdk-ts) · [Python](https://github.com/grignolalouis/logidoc-sdk-python)

## License

MIT
