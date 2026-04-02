# Getting Started

## Prerequisites

- Docker and Docker Compose
- An LLM API key (Anthropic, OpenAI, or any supported provider)

## Install

```bash
git clone https://github.com/grignolalouis/logidoc-server.git
cd logidoc-server
cp .env.example .env
```

Edit `.env` and set your LLM provider:

```env
LLM_PROVIDER=anthropic
LLM_MODEL=claude-haiku-4-5-20251001
LLM_API_KEY=sk-ant-your-key
```

Start the server:

```bash
docker compose up --build
```

## Upload a document

Open `http://localhost:7042` and click **Upload**.

Or via CLI:

```bash
curl -X POST http://localhost:7042/v1/documents -F "file=@my-document.pdf"
```

## Index it

Click **Index** in the UI, or:

```bash
curl -X POST http://localhost:7042/v1/documents/{id}/index
```

The indexation runs in the background. Poll the document status until it's `ready`.

## Connect an agent

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

Your agent now has access to `logidoc_list_documents`, `logidoc_get_toc`, and `logidoc_get_sections`.
