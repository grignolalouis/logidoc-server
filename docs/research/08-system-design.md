# PageIndex System Design — Architecture Complète

## Vue d'ensemble

```
┌──────────────────────────────────────────────────────────────────────┐
│                        REPOS SÉPARÉS                                 │
│                                                                      │
│  ┌─────────────────────┐  ┌──────────────┐  ┌────────────────────┐  │
│  │ pageindex-server     │  │ pageindex-sdk│  │ pageindex-sdk-ts   │  │
│  │ (Go)                 │  │ (Go client)  │  │ (TypeScript client)│  │
│  │                      │  └──────────────┘  └────────────────────┘  │
│  │ • HTTP API           │  ┌──────────────┐                          │
│  │ • MCP Server         │  │ pageindex-sdk│                          │
│  │ • Core Engine        │  │ (Python)     │                          │
│  │ • CLI                │  └──────────────┘                          │
│  └─────────────────────┘                                             │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 1. Repos et packages

### Repo principal: `pageindex-server`

Le serveur complet. Un seul binaire Go qui expose tout.

```
pageindex-server/
├── cmd/
│   └── pageindex/
│       └── main.go              # Entry point (HTTP + MCP)
├── api/
│   └── openapi.yaml             # OpenAPI spec (source de vérité pour les SDKs)
├── internal/
│   ├── engine/                  # Core: parsing, indexing, retrieval
│   │   ├── parser.go            # PDF → pages texte
│   │   ├── ocr.go               # Fallback OCR+LLM pour tables
│   │   ├── indexer.go           # Pages → arbre JSON (appels LLM)
│   │   ├── retriever.go         # Query → navigation arbre → texte extrait
│   │   ├── extractor.go         # Node IDs → texte brut (pas de LLM)
│   │   └── store.go             # Persistence des index JSON
│   ├── httpapi/                 # Handlers HTTP REST
│   │   └── handlers.go
│   └── mcpserver/               # MCP tool definitions + handlers
│       └── server.go
├── pkg/
│   └── types/                   # Types partagés (Document, Node, Index...)
│       └── types.go
├── Dockerfile
└── go.mod
```

### Repos SDK (générés depuis openapi.yaml)

```
pageindex-sdk-go/        # Go client — go get github.com/xxx/pageindex-sdk-go
pageindex-sdk-ts/        # TypeScript — npm install pageindex-sdk
pageindex-sdk-python/    # Python — pip install pageindex-sdk
```

---

## 2. Le serveur — deux interfaces, un engine

### HTTP API (pour les apps/SDKs)

```
POST   /v1/documents              Upload + lance indexation async
GET    /v1/documents              Liste tous les documents
GET    /v1/documents/:id          Metadata + status d'un document
DELETE /v1/documents/:id          Supprime un document + son index
GET    /v1/documents/:id/index    Retourne l'arbre JSON complet
POST   /v1/search                 Recherche dans un ou plusieurs docs
```

#### Détail des endpoints:

**POST /v1/documents**
```
Request:  multipart/form-data { file: binary, name?: string, metadata?: json }
Response: { id: "doc_xxx", name: "report.pdf", status: "indexing", created_at: "..." }
```
→ L'indexation tourne en background. Le client poll le status.

**GET /v1/documents/:id**
```
Response: {
  id: "doc_xxx",
  name: "report.pdf",
  status: "ready" | "indexing" | "error",
  page_count: 142,
  node_count: 38,
  created_at: "...",
  indexed_at: "...",
  error: null
}
```

**POST /v1/search**
```
Request:  { query: "revenue Q3", doc_ids?: ["doc_xxx", "doc_yyy"] }
Response: {
  results: [
    {
      doc_id: "doc_xxx",
      doc_name: "report.pdf",
      node_id: "0012",
      node_title: "Section 3.2: Revenue Breakdown",
      text: "Revenue for Q3 reached $4.2B...",
      score: 0.95
    }
  ]
}
```
→ Si `doc_ids` est omis, cherche dans tous les docs.

### MCP Server (pour les agents LLM)

Utilise `trpc.group/trpc-go/trpc-mcp-go` (la lib MCP du framework trpc-agent-go).
Expose 3 transports: **stdio**, **SSE**, **streamable HTTP**.

```
Tools exposés:

pageindex_search
  Description: "Search indexed documents for information relevant to a query"
  Input:  { query: string, doc_ids?: string[] }
  Output: { results: [{ doc_id, doc_name, node_title, text }] }

pageindex_list_documents
  Description: "List all available indexed documents"
  Input:  {}
  Output: { documents: [{ id, name, status, page_count }] }

pageindex_get_sections
  Description: "Get full text of specific sections from a document"
  Input:  { doc_id: string, node_ids: string[] }
  Output: { sections: [{ node_id, title, text }] }
```

3 tools seulement. L'agent a tout ce qu'il faut.

### Comment les deux partagent le même engine:

```go
// cmd/pageindex/main.go

func main() {
    cfg := loadConfig()

    // Core engine (partagé)
    store := engine.NewFileStore(cfg.DataDir)
    parser := engine.NewPDFParser()
    indexer := engine.NewIndexer(cfg.Model)
    retriever := engine.NewRetriever(cfg.Model)

    core := &engine.Engine{
        Store:     store,
        Parser:    parser,
        Indexer:   indexer,
        Retriever: retriever,
    }

    // HTTP API
    httpServer := httpapi.NewServer(core, cfg.HTTPAddr)
    go httpServer.Start()

    // MCP Server (streamable HTTP sur un autre port, ou même port path différent)
    mcpServer := mcpserver.NewServer(core)
    go mcpServer.Start(cfg.MCPAddr)  // ex: :3001

    // Ou MCP stdio si lancé en mode subprocess
    if cfg.MCPStdio {
        mcpServer.ServeStdio()
    }

    select {} // block forever
}
```

---

## 3. Génération des SDKs

### Stratégie: OpenAPI spec → génération automatique

Le fichier `api/openapi.yaml` est la **source de vérité**. Les SDKs sont générés automatiquement.

### Options de générateurs:

| Outil | Qualité output | Langages | Complexité | Coût |
|-------|---------------|----------|------------|------|
| **openapi-generator** | Correct, verbeux | 50+ | Moyenne | Gratuit |
| **Fern** | Idiomatique, propre | Go, TS, Python, Java... | Simple | Gratuit (open-source) |
| **Speakeasy** | Enterprise-grade | Go, TS, Python... | Simple | Payant |
| **Stainless** | Très propre (utilisé par Anthropic, OpenAI) | TS, Python, Go, Java | Simple | Payant |

### Recommandation: **Fern** ou **openapi-generator**

- **Fern** si on veut des SDKs idiomatiques qui "font pro" sans effort
- **openapi-generator** si on veut rester 100% open-source et gratuit

### Pipeline CI:

```
openapi.yaml modifié → CI génère les SDKs → push dans chaque repo SDK → release
```

### Ce que ça produit (exemple Go):

```go
// pageindex-sdk-go (généré)
package pageindex

type Client struct { ... }

func NewClient(baseURL string, opts ...Option) *Client
func (c *Client) Upload(ctx context.Context, file io.Reader, opts ...UploadOption) (*Document, error)
func (c *Client) List(ctx context.Context) ([]Document, error)
func (c *Client) Get(ctx context.Context, docID string) (*Document, error)
func (c *Client) Delete(ctx context.Context, docID string) error
func (c *Client) Search(ctx context.Context, query string, opts ...SearchOption) (*SearchResult, error)
func (c *Client) GetIndex(ctx context.Context, docID string) (*Index, error)
```

---

## 4. Guide d'intégration utilisateur

### Scénario 1: "Je veux que mon agent puisse lire des PDFs"

**Temps d'intégration: 5 minutes**

```bash
# 1. Lancer le serveur
docker run -p 8080:8080 -p 3001:3001 \
  -e OPENAI_API_KEY=sk-xxx \
  -v ./data:/data \
  pageindex-server
```

```bash
# 2. Indexer un document
curl -X POST http://localhost:8080/v1/documents \
  -F "file=@report.pdf"
# → {"id": "doc_abc", "status": "indexing"}

# Attendre que ce soit prêt
curl http://localhost:8080/v1/documents/doc_abc
# → {"status": "ready", "page_count": 142}
```

```go
// 3. Connecter l'agent au MCP (trpc-agent-go)
mcpToolSet := mcp.NewMCPToolSet(
    mcp.ConnectionConfig{
        Transport: "streamable_http",
        ServerURL: "http://localhost:3001/mcp",
    },
)
mcpToolSet.Init(ctx)

agent := llmagent.New("my-agent",
    llmagent.WithModel(myModel),
    llmagent.WithToolSets([]tool.ToolSet{mcpToolSet}),
)

// L'agent a maintenant accès à: pageindex_search, pageindex_list_documents, pageindex_get_sections
// Il s'en sert automatiquement quand l'utilisateur pose une question sur les docs.
```

**C'est tout. 3 étapes.**

---

### Scénario 2: "Mon app web permet aux users d'uploader des docs et de chatter dessus"

**Temps d'intégration: 30 minutes**

```bash
# 1. Lancer le serveur
docker run -p 8080:8080 -p 3001:3001 -e OPENAI_API_KEY=sk-xxx pageindex-server
```

```typescript
// 2. Backend (Node.js/TypeScript) — utilise le SDK
import { PageIndexClient } from 'pageindex-sdk';

const pageindex = new PageIndexClient('http://localhost:8080');

// Endpoint: user uploade un PDF
app.post('/upload', async (req, res) => {
  const doc = await pageindex.upload(req.file);
  res.json({ docId: doc.id, status: doc.status });
});

// Endpoint: user vérifie le status
app.get('/documents/:id', async (req, res) => {
  const doc = await pageindex.get(req.params.id);
  res.json(doc);
});

// Endpoint: lister les docs disponibles
app.get('/documents', async (req, res) => {
  const docs = await pageindex.list();
  res.json(docs);
});
```

```go
// 3. L'agent (Go) — connecté au même serveur via MCP
// Quand le user envoie un message dans le chat, le backend:
//   a) Récupère les doc_ids attachés à cette session
//   b) Les passe dans le system prompt de l'agent:
//      "L'utilisateur travaille sur les documents suivants: doc_abc, doc_def.
//       Utilise l'outil pageindex_search avec ces doc_ids pour répondre."
//   c) L'agent appelle pageindex_search(query=..., doc_ids=[...]) automatiquement
```

---

### Scénario 3: "J'utilise Claude Desktop / Cursor / n'importe quel client MCP"

**Temps d'intégration: 2 minutes**

```json
// claude_desktop_config.json (ou .cursor/mcp.json)
{
  "mcpServers": {
    "pageindex": {
      "command": "pageindex-server",
      "args": ["--mcp-stdio"],
      "env": {
        "OPENAI_API_KEY": "sk-xxx",
        "DATA_DIR": "/path/to/data"
      }
    }
  }
}
```

Ou via SSE pour un serveur distant:
```json
{
  "mcpServers": {
    "pageindex": {
      "url": "http://my-server:3001/sse"
    }
  }
}
```

L'utilisateur pré-indexe ses docs via CLI ou API, puis Claude/Cursor peut les interroger.

---

### Scénario 4: "Je veux juste indexer des docs en CLI"

```bash
# Indexer
pageindex-server index ./report.pdf
pageindex-server index ./contracts/*.pdf

# Lister
pageindex-server list
# ID           NAME              STATUS   PAGES
# doc_abc      report.pdf        ready    142
# doc_def      contract-a.pdf    ready    38

# Chercher
pageindex-server search "revenue Q3 2025"
# [doc_abc] Section 3.2: Revenue Breakdown (p.45-52)
# Revenue for Q3 reached $4.2B, an increase of 12%...

# Lancer le serveur (HTTP + MCP)
pageindex-server serve --http :8080 --mcp :3001
```

---

## 5. Stack technique

| Composant | Choix | Pourquoi |
|-----------|-------|----------|
| **Langage serveur** | Go | Même écosystème que trpc-agent-go, binaire unique |
| **MCP library** | `trpc-mcp-go` | Déjà utilisée par trpc-agent-go, supporte stdio/SSE/streamable |
| **HTTP framework** | `net/http` stdlib (ou chi) | Simple, pas de dépendance lourde |
| **PDF parsing** | `pdfcpu` ou appel externe `pdftotext` | Go natif ou robuste via poppler |
| **OCR fallback** | API externe (Mistral OCR / vision model) | Pas d'OCR en Go natif qui vaille le coup |
| **LLM calls** | `model.Model` interface de trpc-agent-go | Provider-agnostic |
| **Storage index** | Filesystem JSON (V1) | Simple, portable, pas de DB requise |
| **SDK generation** | OpenAPI spec + Fern ou openapi-generator | Idiomatique, multi-langage, automatisé |
| **Container** | Docker | Déploiement universel |

---

## 6. Résumé — Ce que chaque acteur touche

```
L'UTILISATEUR FINAL (chat):
  → Ne sait pas que PageIndex existe
  → Upload un PDF, pose des questions, obtient des réponses

L'INTÉGRATEUR (dev backend):
  → docker run pageindex-server
  → npm install pageindex-sdk (ou go get / pip install)
  → client.Upload(), client.List(), client.Get()
  → Connecte son agent au MCP (une URL)
  → Passe les doc_ids dans le contexte de l'agent si besoin

L'AGENT (LLM):
  → Voit 3 MCP tools: search, list_documents, get_sections
  → Les appelle quand l'utilisateur pose une question sur les docs
  → Ne sait rien de HTTP, d'upload, ou de stockage

LE SERVEUR PAGEINDEX:
  → Reçoit uploads HTTP, lance l'indexation
  → Répond aux tools MCP avec les résultats de retrieval
  → Un seul binaire, un seul storage, deux interfaces
```
