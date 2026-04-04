# MCP Setup

## Streamable HTTP

For Claude Code, Cursor, Windsurf, or any remote MCP client — add to your project's `.mcp.json`:

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

Restart your client. The tools appear automatically.

## Stdio

For Claude Desktop or other stdio-based clients, run logidoc in stdio mode:

```env
MCP_STDIO=true
```

Then in your client config:

```json
{
  "mcpServers": {
    "logidoc": {
      "command": "/path/to/logidoc-server",
      "env": {
        "MCP_STDIO": "true",
        "LLM_API_KEY": "...",
        "MONGO_URI": "mongodb://localhost:27017"
      }
    }
  }
}
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_ADDR` | `:7043` | MCP server listen address |
| `MCP_TRANSPORT` | `streamable_http` | Transport mode |
| `MCP_STDIO` | `false` | Enable stdio mode |
| `MCP_PORT` | `7043` | Docker host port mapping |
