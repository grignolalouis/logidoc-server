# logidoc

Document indexing for AI agents. No vectors, no embeddings.

Upload a PDF, get a searchable table of contents. Agents browse sections by title and summary, then retrieve the full text of what they need.

## Setup

```bash
git clone https://github.com/grignolalouis/logidoc-server.git
cd logidoc-server
cp .env.example .env
# Set LLM_PROVIDER and LLM_API_KEY in .env
docker compose up --build
```

## MCP

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

## SDKs

Integrate logidoc into your app.

- [Go](https://github.com/grignolalouis/logidoc-sdk-go)
- [TypeScript](https://github.com/grignolalouis/logidoc-sdk-ts)
- [Python](https://github.com/grignolalouis/logidoc-sdk-python)

## Docs

- [Configuration](.env.example)
- [API Reference](api/openapi.yaml)
- [Architecture](docs/research/09-architecture-hexagonal.md)

## License

MIT
