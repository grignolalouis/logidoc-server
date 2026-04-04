# Logidoc Server

Self-hosted document indexing for AI agents. Indexes PDF, DOCX, PPTX, HTML, EPUB into searchable hierarchical trees. Agents browse the table of contents, pick sections, retrieve text. No vectors, no embeddings.

## Quick start

```bash
cp .env.example .env    # set LLM_PROVIDER, LLM_MODEL, LLM_API_KEY
docker compose up --build
```

- HTTP API + UI: `http://localhost:7042`
- MCP server: `http://localhost:7043/mcp`
- Health check: `GET /health`

## Configuration

All via environment variables. See `.env.example` for full list.

**Required:**
- `LLM_PROVIDER` — anthropic, openai, mistral, xai, groq, ollama
- `LLM_MODEL` — model name (e.g. claude-haiku-4-5-20251001, gpt-4o-mini)
- `LLM_API_KEY` — provider API key

**Optional:**
- `VISION_PROVIDER` / `VISION_MODEL` — separate model for image description and table extraction
- `API_KEY` — protect /v1/* endpoints with Bearer token auth
- `INDEXER_ENABLE_TABLE_EXTRACTION` — VLM fallback for garbled tables
- `INDEXER_ENABLE_IMAGE_DESCRIPTION` — extract and describe images via VLM
- `MCP_STDIO=true` — stdio mode for Claude Desktop

## MCP integration

Add to `.mcp.json`:

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

## API usage

```bash
# Upload
curl -X POST http://localhost:7042/v1/documents -F "file=@report.pdf"

# Index
curl -X POST http://localhost:7042/v1/documents/{id}/index

# Get TOC
curl http://localhost:7042/v1/documents/{id}/toc

# Get sections
curl "http://localhost:7042/v1/documents/{id}/sections?ids=chapter-1,section-2"

# Search across all documents
curl "http://localhost:7042/v1/search?q=authentication"
```

## How indexation works

```
Upload → Parse (pdftotext / pandoc) → Detect TOC → Calibrate pages → Build tree → Enrich (tables, images) → Save
```

## Commands

```bash
make dev           # Docker + Air hot reload
make dev-local     # Run without Docker (needs poppler-utils, pandoc)
make build         # Build binary
make test          # Run tests
make prod          # Production Docker deployment
```

## Documentation

Full docs: https://grignolalouis.github.io/logidoc-server/
