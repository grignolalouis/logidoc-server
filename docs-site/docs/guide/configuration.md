# Configuration

All configuration is done via environment variables. Copy `.env.example` to `.env` and edit.

## Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:7042` | HTTP server listen address |
| `MCP_ADDR` | `:7043` | MCP server listen address |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `MONGO_DATABASE` | `logidoc` | Database name |
| `LLM_PROVIDER` | `openai` | LLM provider name |
| `LLM_MODEL` | `gpt-4o` | Model name |
| `LLM_API_KEY` | *required* | Provider API key |
| `LLM_BASE_URL` | | Custom endpoint for OpenAI-compatible providers |
| `API_KEY` | | Optional API key to protect `/v1/*` endpoints |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json` or `text` |

## Docker Compose ports

These variables control the Docker port mapping:

| Variable | Default |
|----------|---------|
| `HTTP_PORT` | `7042` |
| `MCP_PORT` | `7043` |
| `MONGO_PORT` | `27042` |
| `MONGO_VERSION` | `7` |

::: tip
`HTTP_PORT` and `HTTP_ADDR` must use the same port number. Same for `MCP_PORT` and `MCP_ADDR`.
:::
