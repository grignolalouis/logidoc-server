<p align="center">
  <img src=".github/banner.png" alt="logidoc" width="100%">
</p>

<p align="center">
  Document indexing for AI agents. No vectors, no embeddings.
</p>

---

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

- [Go](https://github.com/grignolalouis/logidoc-sdk-go)
- [TypeScript](https://github.com/grignolalouis/logidoc-sdk-ts)
- [Python](https://github.com/grignolalouis/logidoc-sdk-python)

## Documentation

[grignolalouis.github.io/logidoc-server](https://grignolalouis.github.io/logidoc-server/)

## License

MIT
