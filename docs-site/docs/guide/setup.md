# Setup

## Install

```bash
git clone https://github.com/grignolalouis/logidoc-server.git
cd logidoc-server
cp .env.example .env
```

Edit `.env`:

```env
LLM_PROVIDER=anthropic
LLM_MODEL=claude-haiku-4-5-20251001
LLM_API_KEY=sk-ant-your-key
```

Start:

```bash
docker compose up --build
```

UI: `http://localhost:7042` — MCP: `http://localhost:7043/mcp`

## LLM Providers

Set `LLM_PROVIDER` and `LLM_MODEL` in `.env`:

| Provider | `LLM_PROVIDER` | `LLM_MODEL` example |
|----------|----------------|---------------------|
| Anthropic | `anthropic` | `claude-haiku-4-5-20251001` |
| OpenAI | `openai` | `gpt-4o-mini` |
| Mistral | `mistral` | `mistral-large-latest` |
| xAI | `xai` | `grok-2` |
| Groq | `groq` | `llama-3.3-70b-versatile` |
| Ollama | `ollama` | `llama3.2` |

Custom OpenAI-compatible endpoint: set `LLM_BASE_URL`.

## MCP

Add to your project's `.mcp.json`:

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

For stdio mode (Claude Desktop), set `MCP_STDIO=true` in `.env`.

## API Key Auth

Optional. Set `API_KEY` in `.env` to protect `/v1/*` endpoints:

```env
API_KEY=sk-logidoc-your-secret
```

Then pass `Authorization: Bearer sk-logidoc-your-secret` in requests.

Health (`/health`), version (`/version`), and UI (`/ui`) are always public.

## Ports

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:7042` | HTTP API + UI |
| `MCP_ADDR` | `:7043` | MCP server |
| `HTTP_PORT` | `7042` | Docker host mapping |
| `MCP_PORT` | `7043` | Docker host mapping |
| `MONGO_PORT` | `27042` | MongoDB host mapping |

## All Environment Variables

| Variable | Default | Required |
|----------|---------|----------|
| `LLM_PROVIDER` | `openai` | yes |
| `LLM_MODEL` | `gpt-4o` | yes |
| `LLM_API_KEY` | — | **yes** |
| `LLM_BASE_URL` | — | no |
| `MONGO_URI` | `mongodb://localhost:27017` | no |
| `MONGO_DATABASE` | `logidoc` | no |
| `API_KEY` | — | no |
| `LOG_LEVEL` | `info` | no |
| `LOG_FORMAT` | `json` | no |
| `HTTP_RATE_LIMIT` | `100` | no |
| `HTTP_BODY_LIMIT_MB` | `50` | no |
