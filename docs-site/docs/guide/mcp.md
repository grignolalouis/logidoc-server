# MCP Integration

logidoc exposes an MCP server that any compatible client can connect to.

## Streamable HTTP (default)

For Claude Code, Cursor, or any remote client:

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

Place this in `.mcp.json` at the root of your project.

## Stdio (Claude Desktop)

For Claude Desktop or other stdio-based clients, run logidoc in stdio mode:

```env
MCP_STDIO=true
```

Then configure Claude Desktop:

```json
{
  "mcpServers": {
    "logidoc": {
      "command": "/path/to/logidoc-server",
      "env": {
        "MCP_STDIO": "true",
        "LLM_API_KEY": "sk-...",
        "MONGO_URI": "mongodb://localhost:27017"
      }
    }
  }
}
```

## Available tools

| Tool | Parameters | Description |
|------|-----------|-------------|
| `logidoc_list_documents` | — | List all indexed documents |
| `logidoc_get_toc` | `doc_id` | Get the table of contents |
| `logidoc_get_sections` | `doc_id`, `node_ids` | Get full text of sections |

## How agents use it

A typical agent workflow:

1. Call `logidoc_list_documents` to see available documents
2. Call `logidoc_get_toc` to read the structure of a document
3. Reason about which sections are relevant to the user's question
4. Call `logidoc_get_sections` with the relevant node IDs
5. Answer using the retrieved text, citing section titles and page numbers
