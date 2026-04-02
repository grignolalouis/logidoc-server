# Logidoc — Architecture Hexagonale Complète

> **logidoc** = logos (λόγος, raison/logique) + document
> Vectorless, reasoning-based document indexing for AI agents. Open source.

## Stack technique

| Composant | Choix | Import / Version | Pourquoi |
|-----------|-------|-------------------|----------|
| **HTTP framework** | Fiber v3 | `github.com/gofiber/fiber/v3` | Performant, API clean, middleware riche |
| **MCP server** | trpc-mcp-go | `trpc.group/trpc-go/trpc-mcp-go` | Même écosystème, stdio/SSE/streamable |
| **LLM calls** | trpc-agent-go model | `trpc.group/trpc-go/trpc-agent-go/model` | Interface `model.Model` provider-agnostic |
| **OpenAI provider** | trpc-agent-go | `trpc.group/trpc-go/trpc-agent-go/model/openai` | Déjà implémenté, streaming, tool calling |
| **Anthropic provider** | trpc-agent-go | `trpc.group/trpc-go/trpc-agent-go/model/anthropic` | Idem |
| **Logger** | slog (stdlib) | `log/slog` | Standard Go, zero dépendance, structuré |
| **MongoDB** | mongo-driver | `go.mongodb.org/mongo-driver/v2` | Officiel, schemaless, JSON natif |
| **PDF parsing** | pdfcpu ou poppler | `github.com/pdfcpu/pdfcpu` / `pdftotext` | Go natif ou robuste via appel externe |
| **Config** | godotenv | `github.com/joho/godotenv` | Simple, .env based |
| **SDK generation** | Fern | `fern-api` CLI | Idiomatique, multi-langage, CI/CD |
| **Container** | Docker | - | Déploiement universel |

## Vue d'ensemble

```
Ports & Adapters / Hexagonal Architecture
Domain-Driven Design au centre

┌─────────────────────────────────────────────────────────────────────┐
│                            cmd/                                      │
│                         main.go (graceful shutdown, signal handling)  │
│                            │                                         │
│                          app.go (DI container, wiring)               │
│                            │                                         │
│    ┌───────────────────────┼───────────────────────┐                 │
│    │                       │                       │                 │
│    ▼                       ▼                       ▼                 │
│ ┌──────────┐      ┌──────────────┐      ┌────────────────┐          │
│ │ Primary   │      │    Core      │      │  Secondary     │          │
│ │ Adapters  │─────►│              │◄─────│  Adapters      │          │
│ │           │      │  Domain      │      │                │          │
│ │ • HTTP    │      │  Service     │      │ • MongoDB repo │          │
│ │ • MCP     │      │  Port        │      │ • LLM adapter  │          │
│ │           │      │              │      │ • OCR adapter  │          │
│ │           │      │              │      │ • PDF parser   │          │
│ │           │      │              │      │ • Storage      │          │
│ └──────────┘      └──────────────┘      └────────────────┘          │
│                                                                      │
│ ┌──────────────────────────────────────────────────────────────┐     │
│ │                    Infrastructure                             │     │
│ │  logger • config • eventbus • async • crypto • errors        │     │
│ └──────────────────────────────────────────────────────────────┘     │
│                                                                      │
│ ┌──────────────────────────────────────────────────────────────┐     │
│ │                    Test Utilities                              │     │
│ │  mocks • fixtures • helpers                                   │     │
│ └──────────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Arborescence complète

```
logidoc-server/
│
├── cmd/
│   └── server/
│       └── main.go                         # Entry point, graceful shutdown
│
├── app.go                                  # DI wiring, builds the app
│
├── config/
│   ├── config.go                           # Config struct principale
│   ├── loader.go                           # Chargement: env → file → defaults
│   ├── http.go                             # HTTP server config (port, timeouts, cors)
│   ├── mcp.go                              # MCP server config (transport, port)
│   ├── mongo.go                            # MongoDB config (uri, db name, timeouts)
│   ├── llm.go                              # LLM config (provider, model, api key)
│   ├── ocr.go                              # OCR config (provider, api key)
│   ├── logger.go                           # Logger config (level, format, output)
│   └── indexer.go                          # Indexer config (max pages/node, max tokens/node)
│
├── core/
│   ├── domain/
│   │   ├── document.go                     # Document entity
│   │   ├── index.go                        # Index, Node, Tree value objects
│   │   ├── search.go                       # SearchQuery, SearchResult value objects
│   │   └── errors.go                       # Domain errors (ErrDocNotFound, ErrIndexFailed...)
│   │
│   ├── port/
│   │   ├── primary.go                      # Ports driven par l'extérieur (use cases)
│   │   └── secondary.go                    # Ports que le core drive (repos, services externes)
│   │
│   └── service/
│       ├── document_service.go             # Orchestration: upload, status, list, delete
│       ├── indexer_service.go              # Orchestration: parsing → tree building → storage
│       └── retrieval_service.go            # Orchestration: navigation → extraction → résultat
│
├── adapter/
│   ├── primary/
│   │   ├── http/
│   │   │   ├── server.go                   # Fiber v3 app struct, Start(), Shutdown()
│   │   │   ├── router.go                   # Route definitions
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go                 # API key validation
│   │   │   │   ├── cors.go                 # CORS config
│   │   │   │   ├── ratelimit.go            # Rate limiting
│   │   │   │   ├── requestid.go            # Request ID injection
│   │   │   │   └── recovery.go             # Panic recovery
│   │   │   ├── handler/
│   │   │   │   ├── document_handler.go     # Upload, List, Get, Delete handlers
│   │   │   │   └── search_handler.go       # Search handler
│   │   │   ├── request/
│   │   │   │   ├── upload.go               # Upload request validation
│   │   │   │   └── search.go               # Search request validation
│   │   │   └── response/
│   │   │       ├── document.go             # Document response DTOs
│   │   │       ├── search.go               # Search response DTOs
│   │   │       └── error.go                # Error response format
│   │   │
│   │   └── mcp/
│   │       ├── server.go                   # MCP server struct, Start(), Shutdown()
│   │       └── tools.go                    # Tool definitions + handlers (search, list, get_sections)
│   │
│   └── secondary/
│       ├── repository/
│       │   └── mongo/
│       │       ├── document_repo.go        # DocumentRepository impl
│       │       ├── index_repo.go           # IndexRepository impl
│       │       └── connection.go           # Mongo client lifecycle
│       │
│       ├── llm/
│       │   ├── adapter.go                  # LLMProvider port impl
│       │   ├── openai.go                   # OpenAI impl
│       │   └── anthropic.go                # Anthropic impl (optionnel)
│       │
│       └── llm/                            # UNIQUEMENT le LLM adapter pour la retrieval
│                                           #
│                                           # NOTE: PDF parsing et OCR → réutiliser directement
│                                           # depuis trpc-agent-go (pas de secondary adapter custom):
│                                           #   knowledge/document/reader/pdf (pdfcpu)
│                                           #   knowledge/ocr/tesseract
│                                           #   model.Message.AddImageData() pour vision VLM
│
├── infrastructure/
│   ├── logger/
│   │   └── logger.go                       # slog setup (handler, level, format)
│   ├── errors/
│   │   └── errors.go                       # App-level error types, wrapping
│   ├── eventbus/
│   │   └── eventbus.go                     # Async event bus (indexation complete, etc.)
│   ├── async/
│   │   └── worker.go                       # Background worker pool (indexation jobs)
│   └── validator/
│       └── validator.go                    # Request validation helpers
│
├── testutil/
│   ├── mocks/
│   │   ├── document_repo_mock.go           # Mock DocumentRepository
│   │   ├── index_repo_mock.go              # Mock IndexRepository
│   │   └── llm_mock.go                     # Mock LLMProvider
│   ├── fixtures/
│   │   ├── documents.go                    # Test documents
│   │   ├── indexes.go                      # Test index trees
│   │   └── testdata/
│   │       ├── sample.pdf                  # Small test PDF
│   │       └── sample_index.json           # Pre-built test index
│   └── helpers/
│       ├── mongo.go                        # Test MongoDB container (testcontainers)
│       └── assertions.go                   # Custom test assertions
│
├── api/
│   └── openapi.yaml                        # OpenAPI spec (source de vérité SDKs)
│
├── fern/
│   ├── fern.config.json                    # Fern org + version
│   ├── generators.yml                      # SDK generators config (Go, TS, Python)
│   └── openapi/
│       └── openapi.yaml                    # Symlink ou copie → ../api/openapi.yaml
│
├── .env.example                            # Template des variables d'environnement
├── Dockerfile
├── docker-compose.yml                      # Server + MongoDB
├── Makefile
└── go.mod                                  # deps: fiber/v3, trpc-mcp-go, trpc-agent-go, mongo-driver, godotenv
```

---

## Core — Le domaine

### domain/document.go

```go
package domain

import "time"

type DocumentStatus string

const (
    StatusPending  DocumentStatus = "pending"
    StatusIndexing DocumentStatus = "indexing"
    StatusReady    DocumentStatus = "ready"
    StatusError    DocumentStatus = "error"
)

type Document struct {
    ID          string
    Name        string
    Description string          // LLM-generated at indexing
    Status      DocumentStatus
    PageCount   int
    NodeCount   int
    Error       string          // Si status == error
    CreatedAt   time.Time
    IndexedAt   *time.Time
}
```

### domain/index.go

```go
package domain

type Index struct {
    DocID     string
    Tree      []Node
    Version   int             // Schema version pour migration lazy
}

type Node struct {
    ID       string           // "0001", kebab-case, etc.
    Title    string
    Summary  string           // 15-25 mots, LLM-generated
    Text     string           // Texte brut complet de la section
    Children []Node
}
```

### domain/search.go

```go
package domain

type SearchQuery struct {
    Query  string
    DocIDs []string           // Optionnel: scope à certains docs
}

type SearchResult struct {
    Items []SearchResultItem
}

type SearchResultItem struct {
    DocID     string
    DocName   string
    NodeID    string
    NodeTitle string
    Text      string
}
```

### domain/errors.go

```go
package domain

import "errors"

var (
    ErrDocumentNotFound = errors.New("document not found")
    ErrIndexNotFound    = errors.New("index not found for document")
    ErrDocumentNotReady = errors.New("document is not yet indexed")
    ErrInvalidDocument  = errors.New("invalid document format")
    ErrIndexingFailed   = errors.New("indexing failed")
)
```

---

## Core — Les ports

### port/primary.go — Ce que l'extérieur peut demander au core

```go
package port

import (
    "context"
    "io"
    "logidoc-server/core/domain"
)

// DocumentService — port primaire pour la gestion de documents
type DocumentService interface {
    Upload(ctx context.Context, filename string, file io.Reader) (*domain.Document, error)
    Get(ctx context.Context, id string) (*domain.Document, error)
    List(ctx context.Context) ([]domain.Document, error)
    Delete(ctx context.Context, id string) error
}

// RetrievalService — port primaire pour la recherche
type RetrievalService interface {
    Search(ctx context.Context, query domain.SearchQuery) (*domain.SearchResult, error)
    GetSections(ctx context.Context, docID string, nodeIDs []string) ([]domain.Node, error)
}
```

### port/secondary.go — Ce dont le core a besoin de l'extérieur

```go
package port

import (
    "context"
    "io"
    "logidoc-server/core/domain"
)

// DocumentRepository — persistence des documents
type DocumentRepository interface {
    Save(ctx context.Context, doc *domain.Document) error
    FindByID(ctx context.Context, id string) (*domain.Document, error)
    FindAll(ctx context.Context) ([]domain.Document, error)
    UpdateStatus(ctx context.Context, id string, status domain.DocumentStatus, err string) error
    Delete(ctx context.Context, id string) error
}

// IndexRepository — persistence des index
type IndexRepository interface {
    Save(ctx context.Context, index *domain.Index) error
    FindByDocID(ctx context.Context, docID string) (*domain.Index, error)
    Delete(ctx context.Context, docID string) error
}

// NOTE: PDFParser et OCR sont DÉJÀ dans trpc-agent-go, pas besoin de ports custom:
//
//   PDF:  knowledge/document/reader/pdf  (pdfcpu + ledongthuc/pdf)
//         → reader.ReadFromFile(), ReadFromReader(), ReadFromURL()
//         → Extraction texte page par page
//         → Extraction images intégrée via pdfcpu
//         → Option: reader.WithChunk(false) pour avoir les pages brutes
//
//   OCR:  knowledge/ocr/tesseract
//         → ocr.Extractor interface: ExtractText(ctx, imageData) (string, error)
//         → S'injecte dans le PDF reader: reader.WithOCRExtractor(tesseract)
//         → Le reader détecte auto les images et appelle l'OCR
//         → Build tag: -tags tesseract
//
//   Vision/Multimodal: model.Message supporte les images nativement
//         → msg.AddImageData(data, "high", "png")
//         → msg.AddImageURL(url, "auto")
//         → Tous les providers (OpenAI, Anthropic, Gemini) supportent la vision
//         → Alternative à Tesseract: envoyer l'image au VLM directement
//
// Seul port custom nécessaire pour la retrieval:

// LLMProvider — navigation dans l'arbre pour la retrieval
// NOTE: L'indexation utilise un ChainAgent trpc-agent-go (voir 10-pipeline-with-agents.md)
type LLMProvider interface {
    NavigateTree(ctx context.Context, query string, tree []domain.Node) ([]string, error)
}
```

---

## Core — Les services

### service/document_service.go

```go
package service

// DocumentServiceImpl orchestre upload + lance l'indexation async
type DocumentServiceImpl struct {
    docRepo   port.DocumentRepository
    indexRepo port.IndexRepository
    indexer   *IndexerService          // Pour lancer l'indexation
    worker    *async.WorkerPool        // Background jobs
}

// Upload:
//   1. Génère un ID
//   2. Sauve le document en "pending" dans le repo
//   3. Dispatch un job d'indexation async
//   4. Retourne le document immédiatement
```

### service/indexer_service.go

```go
package service

import (
    "trpc.group/trpc-go/trpc-agent-go/agent/chainagent"
    "trpc.group/trpc-go/trpc-agent-go/agent/llmagent"
    "trpc.group/trpc-go/trpc-agent-go/model"
    "trpc.group/trpc-go/trpc-agent-go/runner"
    "trpc.group/trpc-go/trpc-agent-go/session/inmemory"
    "trpc.group/trpc-go/trpc-agent-go/tool/function"
)

// IndexerService utilise un ChainAgent trpc-agent-go pour orchestrer le pipeline
type IndexerService struct {
    pipeline agent.Agent           // ChainAgent: TOC → Parse → Tree → Summary
    runner   runner.Runner         // Runner one-shot (inmemory session)
    docRepo  port.DocumentRepository
    idxRepo  port.IndexRepository
    logger   *slog.Logger
}

// Le pipeline est un ChainAgent composé de 4 LLMAgents:
//
//   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
//   │ TOC Detector  │───►│ Page Parser   │───►│ Tree Builder  │───►│ Summarizer   │
//   │ (structured   │    │ (tools:       │    │ (structured   │    │ (structured  │
//   │  output JSON) │    │  parse_pages  │    │  output JSON) │    │  output JSON)│
//   │               │    │  ocr_page)    │    │               │    │              │
//   └──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
//
// - FunctionTools wrappent le PDFParser et OCRProvider (secondary ports)
// - Chaque agent utilise WithStructuredOutputJSON() pour du JSON validé
// - Le Runner utilise inmemory.SessionService (one-shot, pas de persistence)
//
// Voir: research/pageindex/10-pipeline-with-agents.md pour le détail complet
```

### service/retrieval_service.go

```go
package service

// RetrievalServiceImpl orchestre la recherche
type RetrievalServiceImpl struct {
    docRepo   port.DocumentRepository
    indexRepo port.IndexRepository
    llm       port.LLMProvider
}

// Search(query):
//   1. Résoudre les docs cibles (query.DocIDs ou tous)
//   2. Pour chaque doc:
//      a. Charger l'index (tree avec titres + summaries seulement)
//      b. llm.NavigateTree(query, tree) → nodeIDs
//      c. Extraire le texte des nodes sélectionnés
//   3. Agréger et retourner les résultats
```

---

## Adapters — Primary

### HTTP (adapter/primary/http/)

```go
// server.go
import (
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/cors"
    "github.com/gofiber/fiber/v3/middleware/recover"
    "github.com/gofiber/fiber/v3/middleware/requestid"
    "github.com/gofiber/fiber/v3/middleware/limiter"
)

type Server struct {
    app     *fiber.App
    cfg     config.HTTPConfig
    docH    *handler.DocumentHandler
    searchH *handler.SearchHandler
    logger  *slog.Logger
}

func NewServer(cfg config.HTTPConfig, docSvc port.DocumentService, retSvc port.RetrievalService, logger *slog.Logger) *Server {
    app := fiber.New(fiber.Config{
        ReadTimeout:  cfg.ReadTimeout,
        WriteTimeout: cfg.WriteTimeout,
        ErrorHandler: customErrorHandler,
    })
    s := &Server{app: app, cfg: cfg, logger: logger}
    s.docH = handler.NewDocumentHandler(docSvc)
    s.searchH = handler.NewSearchHandler(retSvc)
    s.setupRoutes()
    return s
}

func (s *Server) Start() error { return s.app.Listen(s.cfg.Addr) }
func (s *Server) Shutdown(ctx context.Context) error { return s.app.ShutdownWithContext(ctx) }

// router.go
func (s *Server) setupRoutes() {
    // Middleware globaux
    s.app.Use(requestid.New())
    s.app.Use(recover.New())
    s.app.Use(cors.New(cors.Config{AllowOrigins: s.cfg.CORSOrigins}))
    s.app.Use(limiter.New(limiter.Config{Max: s.cfg.RateLimit}))

    // Routes v1
    api := s.app.Group("/v1")
    api.Post("/documents", s.docH.Upload)
    api.Get("/documents", s.docH.List)
    api.Get("/documents/:id", s.docH.Get)
    api.Delete("/documents/:id", s.docH.Delete)
    api.Post("/search", s.searchH.Search)
}

// handler/document_handler.go
type DocumentHandler struct {
    svc port.DocumentService      // Dépend du PORT, pas de l'implem
}

func (h *DocumentHandler) Upload(c fiber.Ctx) error {     // Fiber v3: fiber.Ctx (pas *fiber.Ctx)
    file, err := c.FormFile("file")
    if err != nil {
        return c.Status(400).JSON(response.Error("file is required"))
    }
    f, _ := file.Open()
    defer f.Close()
    doc, err := h.svc.Upload(c.Context(), file.Filename, f)
    if err != nil {
        return c.Status(500).JSON(response.Error(err.Error()))
    }
    return c.Status(201).JSON(response.FromDocument(doc))
}
```

### MCP (adapter/primary/mcp/)

```go
// server.go
import (
    mcp "trpc.group/trpc-go/trpc-mcp-go"
)

type Server struct {
    stdio     *mcp.StdioServer   // Mode stdio (Claude Desktop)
    http      *mcp.Server        // Mode streamable HTTP
    cfg       config.MCPConfig
    retSvc    port.RetrievalService
    docSvc    port.DocumentService
    logger    *slog.Logger
}

func NewServer(cfg config.MCPConfig, docSvc port.DocumentService, retSvc port.RetrievalService, logger *slog.Logger) *Server {
    s := &Server{cfg: cfg, docSvc: docSvc, retSvc: retSvc, logger: logger}

    if cfg.Stdio {
        s.stdio = mcp.NewStdioServer("logidoc", "1.0.0",
            mcp.WithStdioServerLogger(mcp.GetDefaultLogger()),
        )
        s.registerTools(s.stdio)
    } else {
        s.http = mcp.NewServer("logidoc", "1.0.0",
            mcp.WithServerAddress(cfg.Addr),
        )
        s.registerTools(s.http)
    }
    return s
}

func (s *Server) Start() error {
    if s.cfg.Stdio {
        return s.stdio.Serve()
    }
    return s.http.Serve()
}

func (s *Server) Shutdown(ctx context.Context) error {
    // Cleanup
    return nil
}

// tools.go
// registerTools fonctionne avec les deux types de server (interface commune)
func (s *Server) registerTools(srv interface{ RegisterTool(mcp.Tool, any) }) {

    // Tool 1: search
    searchTool := mcp.NewTool("logidoc_search",
        mcp.WithDescription("Search indexed documents for information relevant to a query"),
        mcp.WithString("query", mcp.Required(), mcp.Description("The search query")),
        mcp.WithString("doc_ids", mcp.Description("Optional: comma-separated document IDs to scope search")),
    )
    srv.RegisterTool(searchTool, s.handleSearch)

    // Tool 2: list_documents
    listTool := mcp.NewTool("logidoc_list_documents",
        mcp.WithDescription("List all indexed documents with titles, descriptions and top-level sections"),
    )
    srv.RegisterTool(listTool, s.handleListDocuments)

    // Tool 3: get_sections
    sectionsTool := mcp.NewTool("logidoc_get_sections",
        mcp.WithDescription("Get full text of specific sections from a document by node IDs"),
        mcp.WithString("doc_id", mcp.Required(), mcp.Description("Document ID")),
        mcp.WithString("node_ids", mcp.Required(), mcp.Description("Comma-separated node IDs to retrieve")),
    )
    srv.RegisterTool(sectionsTool, s.handleGetSections)
}

// handlers — appellent les mêmes ports que les HTTP handlers
func (s *Server) handleSearch(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    query := req.Params.Arguments["query"].(string)
    result, err := s.retSvc.Search(ctx, domain.SearchQuery{Query: query})
    if err != nil {
        return nil, err
    }
    // Sérialiser en texte lisible par le LLM
    text := formatSearchResults(result)
    return mcp.NewTextResult(text), nil
}

func (s *Server) handleListDocuments(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    docs, err := s.docSvc.List(ctx)
    if err != nil {
        return nil, err
    }
    text := formatDocumentList(docs)
    return mcp.NewTextResult(text), nil
}
```

---

## Adapters — Secondary

### MongoDB (adapter/secondary/repository/mongo/)

```go
// connection.go
type Connection struct {
    client *mongo.Client
    db     *mongo.Database
}

func NewConnection(cfg config.MongoConfig) (*Connection, error)
func (c *Connection) Close(ctx context.Context) error

// document_repo.go
type DocumentRepo struct {
    collection *mongo.Collection
}

func (r *DocumentRepo) Save(ctx context.Context, doc *domain.Document) error {
    _, err := r.collection.InsertOne(ctx, toMongo(doc))
    return err
}
// ... FindByID, FindAll, UpdateStatus, Delete

// index_repo.go — l'arbre JSON est stocké tel quel dans Mongo
type IndexRepo struct {
    collection *mongo.Collection
}

func (r *IndexRepo) Save(ctx context.Context, index *domain.Index) error {
    // Upsert: remplace si existe déjà
    _, err := r.collection.ReplaceOne(ctx,
        bson.M{"doc_id": index.DocID},
        toMongo(index),
        options.Replace().SetUpsert(true),
    )
    return err
}
```

### LLM (adapter/secondary/llm/) — UNIQUEMENT pour la retrieval

```go
// adapter.go — implémente port.LLMProvider pour la navigation dans l'arbre
// NOTE: L'indexation utilise un ChainAgent (pas cet adapter)
import (
    agentmodel "trpc.group/trpc-go/trpc-agent-go/model"
    "trpc.group/trpc-go/trpc-agent-go/model/openai"
    "trpc.group/trpc-go/trpc-agent-go/model/anthropic"
)

type Adapter struct {
    model  agentmodel.Model    // Interface model.Model de trpc-agent-go
    logger *slog.Logger
}

func NewAdapter(cfg config.LLMConfig, logger *slog.Logger) *Adapter {
    var m agentmodel.Model
    switch cfg.Provider {
    case "openai":
        m = openai.New(cfg.Model, openai.WithAPIKey(cfg.APIKey))
    case "anthropic":
        m = anthropic.New(cfg.Model, anthropic.WithAPIKey(cfg.APIKey))
    default:
        m = openai.New(cfg.Model)
    }
    return &Adapter{model: m, logger: logger}
}

func (a *Adapter) NavigateTree(ctx context.Context, query string, tree []domain.Node) ([]string, error) {
    treeJSON := serializeTreeCompact(tree)  // titres + summaries seulement
    req := &agentmodel.Request{
        Messages: []agentmodel.Message{
            agentmodel.NewSystemMessage(navigationPrompt),
            agentmodel.NewUserMessage("TOC:\n" + treeJSON + "\n\nQuery: " + query),
        },
        GenerationConfig: agentmodel.GenerationConfig{Stream: false},
    }
    respChan, err := a.model.GenerateContent(ctx, req)
    if err != nil {
        return nil, err
    }
    var result string
    for resp := range respChan {
        if resp.Error != nil {
            return nil, fmt.Errorf("llm error: %s", resp.Error.Message)
        }
        if len(resp.Choices) > 0 {
            result += resp.Choices[0].Message.Content
        }
    }
    return strings.Split(strings.TrimSpace(result), ","), nil
}
```

---

## DI Wiring — app.go

```go
package main

import (
    "context"
    "log/slog"

    "logidoc-server/config"
    "logidoc-server/core/service"
    httpapi "logidoc-server/adapter/primary/http"
    mcpapi "logidoc-server/adapter/primary/mcp"
    mongorepo "logidoc-server/adapter/secondary/repository/mongo"
    "logidoc-server/adapter/secondary/llm"
    // PDF + OCR: directement depuis trpc-agent-go (pas de custom adapter)
    pdfreader "trpc.group/trpc-go/trpc-agent-go/knowledge/document/reader/pdf"
    "trpc.group/trpc-go/trpc-agent-go/knowledge/document/reader"
    "trpc.group/trpc-go/trpc-agent-go/knowledge/ocr/tesseract"
    "logidoc-server/infrastructure/async"
    infralog "logidoc-server/infrastructure/logger"
)

type App struct {
    HTTPServer *httpapi.Server
    MCPServer  *mcpapi.Server
    MongoDB    *mongorepo.Connection
    Worker     *async.WorkerPool
    Logger     *slog.Logger
}

func NewApp(cfg *config.Config) (*App, error) {
    // Logger (slog)
    logger := infralog.New(cfg.Logger)  // retourne *slog.Logger configuré
    slog.SetDefault(logger)

    // Secondary adapters
    mongConn, err := mongorepo.NewConnection(cfg.Mongo)
    if err != nil {
        return nil, fmt.Errorf("mongo connection: %w", err)
    }
    docRepo := mongorepo.NewDocumentRepo(mongConn)
    indexRepo := mongorepo.NewIndexRepo(mongConn)

    // LLM adapter via trpc-agent-go model.Model
    llmAdapter := llm.NewAdapter(cfg.LLM, logger)

    // PDF Reader + OCR — directement depuis trpc-agent-go
    ocrExtractor, _ := tesseract.New(tesseract.WithLanguage("eng"))
    pdfReader := pdfreader.New(
        reader.WithChunk(false),               // Pages brutes, pas de chunking
        reader.WithOCRExtractor(ocrExtractor),  // OCR auto sur les images/tables
    )

    // Infrastructure
    worker := async.NewWorkerPool(cfg.Workers)

    // Core services (ne connaissent que les ports)
    indexerSvc := service.NewIndexerService(docRepo, indexRepo, pdfReader, llmAdapter, cfg.Indexer, logger)
    docSvc := service.NewDocumentService(docRepo, indexRepo, indexerSvc, worker, logger)
    retSvc := service.NewRetrievalService(docRepo, indexRepo, llmAdapter, logger)

    // Primary adapters — HTTP (Fiber v3) + MCP (trpc-mcp-go)
    httpServer := httpapi.NewServer(cfg.HTTP, docSvc, retSvc, logger)
    mcpServer := mcpapi.NewServer(cfg.MCP, docSvc, retSvc, logger)

    return &App{
        HTTPServer: httpServer,
        MCPServer:  mcpServer,
        MongoDB:    mongConn,
        Worker:     worker,
        Logger:     logger,
    }, nil
}

func (a *App) Shutdown(ctx context.Context) error {
    a.Logger.Info("shutting down services...")
    a.Worker.Stop()
    a.HTTPServer.Shutdown(ctx)
    a.MCPServer.Shutdown(ctx)
    a.MongoDB.Close(ctx)
    return nil
}
```

---

## Entry point — cmd/server/main.go

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
    "time"

    "logidoc-server/config"
)

func main() {
    // Config
    cfg, err := config.Load()
    if err != nil {
        slog.Error("failed to load config", "error", err)
        os.Exit(1)
    }

    // Build app (DI)
    app, err := NewApp(cfg)
    if err != nil {
        slog.Error("failed to initialize app", "error", err)
        os.Exit(1)
    }

    // Start HTTP server (Fiber v3)
    go func() {
        if err := app.HTTPServer.Start(); err != nil {
            slog.Error("HTTP server error", "error", err)
            os.Exit(1)
        }
    }()

    // Start MCP server (trpc-mcp-go)
    go func() {
        if err := app.MCPServer.Start(); err != nil {
            slog.Error("MCP server error", "error", err)
            os.Exit(1)
        }
    }()

    slog.Info("Logidoc server started",
        "http", cfg.HTTP.Addr,
        "mcp_transport", cfg.MCP.Transport,
        "mcp_addr", cfg.MCP.Addr,
        "llm_provider", cfg.LLM.Provider,
        "llm_model", cfg.LLM.Model,
    )

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    slog.Info("shutting down...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := app.Shutdown(ctx); err != nil {
        slog.Error("shutdown error", "error", err)
    }
    slog.Info("server stopped")
}
```

---

## Config — config/

### config.go

```go
package config

type Config struct {
    App     AppConfig
    HTTP    HTTPConfig
    MCP     MCPConfig
    Mongo   MongoConfig
    LLM     LLMConfig
    OCR     OCRConfig
    Indexer IndexerConfig
    Logger  LoggerConfig
    Workers int
}

type AppConfig struct {
    Name    string
    Version string
    Env     string  // development, staging, production
}
```

### loader.go

```go
package config

import "github.com/joho/godotenv"

func Load() (*Config, error) {
    // 1. Charger .env si présent (dev local)
    godotenv.Load()

    // 2. Lire les variables d'environnement
    cfg := &Config{
        HTTP: HTTPConfig{
            Addr:         getEnv("HTTP_ADDR", ":8080"),
            ReadTimeout:  getDuration("HTTP_READ_TIMEOUT", 30*time.Second),
            WriteTimeout: getDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
            CORSOrigins:  getEnv("HTTP_CORS_ORIGINS", "*"),
            RateLimit:    getInt("HTTP_RATE_LIMIT", 100),
        },
        MCP: MCPConfig{
            Addr:      getEnv("MCP_ADDR", ":3001"),
            Transport: getEnv("MCP_TRANSPORT", "streamable_http"),
            Stdio:     getBool("MCP_STDIO", false),
        },
        Mongo: MongoConfig{
            URI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
            Database: getEnv("MONGO_DATABASE", "logidoc"),
        },
        LLM: LLMConfig{
            Provider: getEnv("LLM_PROVIDER", "openai"),
            Model:    getEnv("LLM_MODEL", "gpt-4o"),
            APIKey:   requireEnv("LLM_API_KEY"),
        },
        Indexer: IndexerConfig{
            MaxPagesPerNode:  getInt("INDEXER_MAX_PAGES_PER_NODE", 10),
            MaxTokensPerNode: getInt("INDEXER_MAX_TOKENS_PER_NODE", 20000),
            TOCCheckPages:    getInt("INDEXER_TOC_CHECK_PAGES", 20),
        },
        Logger: LoggerConfig{
            Level:  getEnv("LOG_LEVEL", "info"),
            Format: getEnv("LOG_FORMAT", "json"),  // "json" ou "text"
        },
        // ...
    }

    return cfg, cfg.Validate()
}
```

### infrastructure/logger/logger.go

```go
package logger

import (
    "log/slog"
    "os"

    "logidoc-server/config"
)

func New(cfg config.LoggerConfig) *slog.Logger {
    // Parse level
    var level slog.Level
    switch cfg.Level {
    case "debug":
        level = slog.LevelDebug
    case "warn":
        level = slog.LevelWarn
    case "error":
        level = slog.LevelError
    default:
        level = slog.LevelInfo
    }

    opts := &slog.HandlerOptions{Level: level}

    // Choose handler: JSON pour prod, Text pour dev
    var handler slog.Handler
    switch cfg.Format {
    case "text":
        handler = slog.NewTextHandler(os.Stdout, opts)
    default:
        handler = slog.NewJSONHandler(os.Stdout, opts)
    }

    return slog.New(handler)
}
```

### .env.example

```env
# App
APP_ENV=development

# HTTP Server
HTTP_ADDR=:8080

# MCP Server
MCP_ADDR=:3001
MCP_TRANSPORT=streamable_http
# MCP_STDIO=true    # Décommenter pour mode Claude Desktop

# MongoDB
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=logidoc

# LLM
LLM_PROVIDER=openai
LLM_MODEL=gpt-4o
LLM_API_KEY=sk-your-key-here

# OCR (optionnel, pour fallback tables)
# OCR_PROVIDER=openai
# OCR_MODEL=gpt-4o
# OCR_API_KEY=sk-your-key-here

# Indexer
INDEXER_MAX_PAGES_PER_NODE=10
INDEXER_MAX_TOKENS_PER_NODE=20000
INDEXER_TOC_CHECK_PAGES=20

# Logger
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## OpenAPI Spec — api/openapi.yaml

```yaml
openapi: 3.1.0
info:
  title: Logidoc API
  version: 1.0.0
  description: Vectorless, reasoning-based document indexing and retrieval API

servers:
  - url: http://localhost:8080
    description: Local development

paths:
  /v1/documents:
    post:
      operationId: uploadDocument
      summary: Upload and index a document
      requestBody:
        content:
          multipart/form-data:
            schema:
              type: object
              required: [file]
              properties:
                file:
                  type: string
                  format: binary
                name:
                  type: string
                metadata:
                  type: object
      responses:
        '201':
          description: Document accepted for indexing
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Document'

    get:
      operationId: listDocuments
      summary: List all indexed documents
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  documents:
                    type: array
                    items:
                      $ref: '#/components/schemas/DocumentSummary'

  /v1/documents/{id}:
    get:
      operationId: getDocument
      summary: Get document details and status
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Document'
        '404':
          $ref: '#/components/responses/NotFound'

    delete:
      operationId: deleteDocument
      summary: Delete a document and its index
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '204':
          description: Deleted

  /v1/search:
    post:
      operationId: searchDocuments
      summary: Search across indexed documents
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/SearchRequest'
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SearchResponse'

components:
  schemas:
    Document:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        description:
          type: string
        status:
          type: string
          enum: [pending, indexing, ready, error]
        page_count:
          type: integer
        node_count:
          type: integer
        error:
          type: string
          nullable: true
        created_at:
          type: string
          format: date-time
        indexed_at:
          type: string
          format: date-time
          nullable: true

    DocumentSummary:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        description:
          type: string
        status:
          type: string
        page_count:
          type: integer
        top_sections:
          type: array
          items:
            type: string

    SearchRequest:
      type: object
      required: [query]
      properties:
        query:
          type: string
        doc_ids:
          type: array
          items:
            type: string

    SearchResponse:
      type: object
      properties:
        results:
          type: array
          items:
            $ref: '#/components/schemas/SearchResultItem'

    SearchResultItem:
      type: object
      properties:
        doc_id:
          type: string
        doc_name:
          type: string
        node_id:
          type: string
        node_title:
          type: string
        text:
          type: string

  responses:
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            type: object
            properties:
              error:
                type: string
              message:
                type: string
```

---

## Fern SDK Generation — fern/

### fern/fern.config.json

```json
{
  "organization": "logidoc",
  "version": "0.65.32"
}
```

### fern/generators.yml

```yaml
api:
  path: openapi/openapi.yaml

groups:
  sdk-go:
    generators:
      - name: fernapi/fern-go-sdk
        version: 0.0.142
        github:
          repository: logidoc/logidoc-sdk-go

  sdk-typescript:
    generators:
      - name: fernapi/fern-typescript-node-sdk
        version: 0.0.249
        output:
          location: npm
          package-name: "logidoc-sdk"
          token: ${NPM_TOKEN}
        github:
          repository: logidoc/logidoc-sdk-ts

  sdk-python:
    generators:
      - name: fernapi/fern-python-sdk
        version: 0.0.180
        output:
          location: pypi
          package-name: "logidoc-sdk"
          username: __token__
          password: ${PYPI_TOKEN}
        github:
          repository: logidoc/logidoc-sdk-python
```

### Workflow CI (.github/workflows/generate-sdks.yml)

```yaml
name: Generate SDKs
on:
  push:
    paths:
      - 'api/openapi.yaml'
    branches: [main]

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - run: npm install -g fern-api
      - run: fern check
      - run: fern generate --group sdk-go
        env:
          FERN_TOKEN: ${{ secrets.FERN_TOKEN }}
      - run: fern generate --group sdk-typescript
        env:
          FERN_TOKEN: ${{ secrets.FERN_TOKEN }}
          NPM_TOKEN: ${{ secrets.NPM_TOKEN }}
      - run: fern generate --group sdk-python
        env:
          FERN_TOKEN: ${{ secrets.FERN_TOKEN }}
          PYPI_TOKEN: ${{ secrets.PYPI_TOKEN }}
```

---

## Docker Compose — docker-compose.yml

```yaml
version: '3.8'

services:
  logidoc:
    build: .
    ports:
      - "8080:8080"   # HTTP API
      - "3001:3001"   # MCP Server
    environment:
      - MONGO_URI=mongodb://mongo:27017
      - MONGO_DATABASE=logidoc
      - LLM_API_KEY=${LLM_API_KEY}
    depends_on:
      - mongo

  mongo:
    image: mongo:7
    ports:
      - "27017:27017"
    volumes:
      - mongo-data:/data/db

volumes:
  mongo-data:
```

---

## Makefile

```makefile
.PHONY: dev build test lint docker sdk

# Dev
dev:
	go run cmd/server/main.go

# Build
build:
	go build -o bin/logidoc-server cmd/server/main.go

# Test
test:
	go test ./... -v -cover

test-integration:
	go test ./... -v -tags=integration

# Lint
lint:
	golangci-lint run

# Docker
docker:
	docker compose up --build

# SDK generation
sdk:
	cd fern && fern check && fern generate --local

# OpenAPI validation
api-check:
	npx @redocly/cli lint api/openapi.yaml
```

---

## Résumé des repos à créer

```
Repo 1: logidoc-server           ← Ce document décrit ce repo
  Le serveur complet (HTTP + MCP + Core + Config)
  → docker pull / go install

Repo 2: logidoc-sdk-go           ← Généré par Fern
  → go get github.com/logidoc/logidoc-sdk-go

Repo 3: logidoc-sdk-ts           ← Généré par Fern
  → npm install logidoc-sdk

Repo 4: logidoc-sdk-python       ← Généré par Fern
  → pip install logidoc-sdk

Nom: logidoc = logos (λόγος, raison) + document
"Tu connais logidoc? Pour indexer les documents pour tes agents, open source, super simple."
```

Les repos SDK sont **générés automatiquement**. On ne touche jamais au code dedans.
On modifie `api/openapi.yaml` → CI → Fern génère → push dans les repos SDK → release.
