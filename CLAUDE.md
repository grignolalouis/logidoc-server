# Logidoc Server

> **logidoc** = logos (λόγος, raison/logique) + document
> Vectorless, reasoning-based document indexing for AI agents. Open source.

## What is this project?

Logidoc is a self-hosted server that indexes documents (PDF, Markdown) into hierarchical JSON tree structures, then exposes reasoning-based retrieval via MCP tools and HTTP API. Any MCP-compatible agent (trpc-agent-go, Claude Desktop, Cursor, etc.) can search indexed documents without vector databases or chunking.

Inspired by PageIndex methodology (VectifyAI) but implemented as a standalone Go server with clean architecture.

## Architecture

**Hexagonal Architecture / Ports & Adapters with DDD at the center.**

```
cmd/server/main.go              → Entry point, graceful shutdown
app.go                           → DI wiring

core/
  domain/                        → Document, Index, Node, SearchResult, errors
  port/
    primary.go                   → DocumentService, RetrievalService (interfaces)
    secondary.go                 → DocumentRepo, IndexRepo, LLMProvider (interfaces)
  service/                       → Implementations orchestrating the pipeline

adapter/
  primary/
    http/                        → Fiber v3 (server, router, middleware, handler/, request/, response/)
    mcp/                         → trpc-mcp-go (server, tools: search, list_documents, get_sections)
  secondary/
    repository/mongo/            → MongoDB implementations
    llm/                         → LLM adapter for retrieval (navigation only)

infrastructure/                  → logger (slog), errors, eventbus, async worker pool, validator
config/                          → Config structs + godotenv loader
testutil/                        → mocks, fixtures, helpers
api/openapi.yaml                 → OpenAPI spec (source of truth for SDKs)
fern/                            → SDK generation config (Go, TS, Python)
```

## Stack

| Component | Choice | Import |
|-----------|--------|--------|
| HTTP | Fiber v3 | `github.com/gofiber/fiber/v3` |
| MCP server | trpc-mcp-go | `trpc.group/trpc-go/trpc-mcp-go` |
| LLM interface | trpc-agent-go model.Model | `trpc.group/trpc-go/trpc-agent-go/model` |
| LLM providers | OpenAI, Anthropic via trpc-agent-go | `trpc-agent-go/model/openai`, `trpc-agent-go/model/anthropic` |
| PDF parsing | trpc-agent-go knowledge reader | `trpc-agent-go/knowledge/document/reader/pdf` (pdfcpu) |
| OCR | trpc-agent-go tesseract | `trpc-agent-go/knowledge/ocr/tesseract` |
| Vision/multimodal | trpc-agent-go model.Message | `msg.AddImageData()` for VLM fallback |
| Indexation pipeline | trpc-agent-go ChainAgent | `trpc-agent-go/agent/chainagent` + `llmagent` + `function` tools |
| Structured output | trpc-agent-go | `llmagent.WithStructuredOutputJSON()` |
| Logger | slog (stdlib) | `log/slog` |
| Database | MongoDB | `go.mongodb.org/mongo-driver/v2` |
| Config | godotenv | `github.com/joho/godotenv` |

## Key design decisions

1. **JSON is self-contained** — after indexing, the PDF is not needed. Full text is stored in each tree node.
2. **Tables: parser first, OCR+LLM fallback** — pdfcpu extracts text, tesseract OCR handles images, VLM vision as last resort.
3. **Priority = robust pipeline** — don't optimize tokens yet, focus on producing complete and reliable JSON indexes.
4. **MongoDB, no migrations** — schemaless, lazy migration via document version field. `docker pull && restart` = update.
5. **ChainAgent for indexation** — deterministic pipeline (TOC detect → parse → tree build → summarize), not agentic free-form.
6. **model.Model direct for retrieval** — single LLM call to navigate tree, no agent overhead needed.

## Indexation pipeline (ChainAgent)

```
ChainAgent "indexing-pipeline"
  ├── TOC Detector    (LLMAgent + StructuredOutputJSON → TOCDetectionResult)
  ├── Page Parser     (LLMAgent + FunctionTools: parse_document, vision_parse_page)
  ├── Tree Builder    (LLMAgent + StructuredOutputJSON → DocumentTree)
  └── Summarizer      (LLMAgent + StructuredOutputJSON → DocumentMeta)
```

FunctionTools wrap the framework's PDF reader (`knowledge/document/reader/pdf`) and OCR (`knowledge/ocr/tesseract`).

## Retrieval flow

```
Query + TOC (titles+summaries) → LLM NavigateTree → node IDs
Node IDs → extract text from stored JSON → return results
```

## What the server exposes

### HTTP API (for apps/SDKs)
```
POST   /v1/documents              Upload + async indexation
GET    /v1/documents              List all documents
GET    /v1/documents/:id          Get document details + status
DELETE /v1/documents/:id          Delete document + index
POST   /v1/search                 Search across documents
```

### MCP Tools (for LLM agents)
```
pageindex_search(query, doc_ids?)         Search indexed documents
pageindex_list_documents()                List docs with titles, descriptions, top sections
pageindex_get_sections(doc_id, node_ids)  Get full text of specific sections
```

## Research docs

Detailed research and architecture documents are in the sibling repo:
`../trpc-agent-go/research/pageindex/`

- `01-overview.md` — What PageIndex methodology is
- `02-methodology.md` — The 4-phase pipeline
- `03-data-structures.md` — TOCNode JSON schema
- `06-ingestion-deep-dive.md` — Detailed Q&A on ingestion
- `07-design-decisions.md` — Key decisions
- `09-architecture-hexagonal.md` — Full hexagonal architecture with code
- `10-pipeline-with-agents.md` — Pipeline using trpc-agent-go agents

## Commands

```bash
make dev          # Run locally
make build        # Build binary
make test         # Run tests
make docker       # Docker compose (server + mongo)
make sdk          # Generate SDKs via Fern
make api-check    # Validate OpenAPI spec
```
