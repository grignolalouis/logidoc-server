# Logidoc — Pipeline d'indexation avec trpc-agent-go

## Le changement d'approche

Au lieu de faire des appels `model.Model.GenerateContent()` bruts dans un service Go classique,
on utilise le **framework d'agents** de trpc-agent-go pour orchestrer le pipeline.

### Pourquoi c'est mieux:
- **Structured Output** natif → le LLM retourne du JSON validé par schema
- **Tool calling** → le LLM décide quand appeler le parser, l'OCR, etc.
- **ChainAgent** → pipeline séquentiel propre, chaque étape = un agent spécialisé
- **Streaming events** → suivi en temps réel de l'avancement
- **Error handling** intégré (dual-layer: system + API errors)
- **Session management** pour reprendre un indexation interrompue

---

## Architecture du pipeline

### Option A: ChainAgent (pipeline séquentiel fixe)

Chaque étape est un agent spécialisé, exécutés dans l'ordre:

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  TOC Detector │───►│  Page Parser  │───►│ Tree Builder  │───►│  Summarizer  │
│  (LLMAgent)   │    │  (LLMAgent   │    │  (LLMAgent   │    │  (LLMAgent)  │
│               │    │   + tools)   │    │   structured │    │              │
│  Détecte si   │    │   Parse les  │    │   output)    │    │  Génère      │
│  TOC existe   │    │   pages      │    │   Construit  │    │  summaries   │
│  + l'extrait  │    │   OCR tables │    │   l'arbre    │    │  + desc doc  │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
```

```go
pipeline := chainagent.New("indexing-pipeline",
    chainagent.WithSubAgents([]agent.Agent{
        tocDetectorAgent,
        pageParserAgent,
        treeBuilderAgent,
        summarizerAgent,
    }),
)
```

### Option B: LLMAgent + Tools (orchestration par le LLM)

Un seul agent avec des tools, le LLM décide de l'ordre:

```
┌─────────────────────────────────────────┐
│  Indexer Agent (LLMAgent)               │
│                                         │
│  System prompt:                         │
│  "Tu es un indexeur de documents.       │
│   Utilise les tools pour:               │
│   1. Détecter la TOC                    │
│   2. Parser les pages                   │
│   3. Construire l'arbre                 │
│   4. Générer les summaries"             │
│                                         │
│  Tools:                                 │
│  ├── detect_toc(pages) → TOC            │
│  ├── parse_pages(range) → text          │
│  ├── ocr_page(page_num) → text          │
│  ├── build_node(title, pages) → Node    │
│  └── generate_summary(text) → string    │
└─────────────────────────────────────────┘
```

### Recommandation: Option A (ChainAgent)

Pourquoi:
- Le pipeline est **déterministe** — on veut toujours les mêmes étapes dans le même ordre
- Plus **prévisible** et **debuggable** qu'un agent qui décide seul
- Chaque agent a un **structured output** typé → on parse le résultat avec certitude
- Plus **économique** en tokens — pas besoin que le LLM raisonne sur quoi faire

L'Option B (agent + tools) serait meilleure pour la **retrieval** (l'agent navigue l'arbre),
pas pour l'indexation qui est un pipeline fixe.

---

## Détail de chaque agent du pipeline

### Agent 1: TOC Detector

```go
type TOCDetectionResult struct {
    HasTOC    bool      `json:"has_toc"`
    TOCPages  []int     `json:"toc_pages"`    // Numéros de pages contenant la TOC
    Entries   []TOCEntry `json:"entries"`      // Entrées de la TOC extraites
    Mode      string    `json:"mode"`          // "with_pages", "without_pages", "no_toc"
}

type TOCEntry struct {
    Title string `json:"title"`
    Level int    `json:"level"`    // 1 = chapter, 2 = section, etc.
    Page  *int   `json:"page"`     // nil si pas de numéro de page
}

tocAgent := llmagent.New("toc-detector",
    llmagent.WithModel(llm),
    llmagent.WithInstruction(`Analyze the first pages of this document.
Determine if there is a Table of Contents.
If yes, extract all entries with their hierarchy level and page numbers (if available).
Return the mode: "with_pages" if page numbers exist, "without_pages" if not, "no_toc" if no TOC found.`),
    llmagent.WithStructuredOutputJSON(
        (*TOCDetectionResult)(nil),
        true,
        "Table of contents detection result",
    ),
    llmagent.WithGenerationConfig(model.GenerationConfig{
        Temperature: floatPtr(0.1),  // Très déterministe
        Stream:      false,
    }),
)
```

### Agent 2: Page Parser (avec tool wrappant le PDF Reader du framework)

```go
// Le PDF Reader de trpc-agent-go fait TOUT: texte + OCR images/tables
// On le wrappe dans un FunctionTool pour que l'agent puisse l'appeler

import (
    pdfreader "trpc.group/trpc-go/trpc-agent-go/knowledge/document/reader/pdf"
    "trpc.group/trpc-go/trpc-agent-go/knowledge/document/reader"
    "trpc.group/trpc-go/trpc-agent-go/knowledge/ocr/tesseract"
)

// Créer le PDF reader avec OCR intégré (une seule fois au démarrage)
ocrExtractor, _ := tesseract.New(tesseract.WithLanguage("eng"))
pdfReader := pdfreader.New(
    reader.WithChunk(false),               // Pages brutes, pas de chunking
    reader.WithOCRExtractor(ocrExtractor),  // Auto-OCR sur images/tables dans le PDF
)

// Tool: parse le document entier (texte + OCR auto)
parseDocTool := function.NewFunctionTool(
    func(ctx context.Context, input struct {
        FilePath string `json:"file_path"`
    }) ([]map[string]string, error) {
        // Le PDF reader extrait tout: texte + OCR images page par page
        docs, err := pdfReader.ReadFromFile(input.FilePath)
        if err != nil {
            return nil, err
        }
        // Retourner le contenu par document/page
        var pages []map[string]string
        for _, doc := range docs {
            pages = append(pages, map[string]string{
                "name":    doc.Name,
                "content": doc.Content,
            })
        }
        return pages, nil
    },
    function.WithName("parse_document"),
    function.WithDescription("Parse a PDF document extracting text and OCR for tables/images"),
)

// Alternative: Tool vision pour les pages complexes (envoie l'image au VLM)
// Utilise model.Message.AddImageData() pour envoyer une page en image au LLM
visionPageTool := function.NewFunctionTool(
    func(ctx context.Context, input struct {
        PageImageBase64 string `json:"page_image"`
        Question        string `json:"question"`
    }) (string, error) {
        msg := model.NewUserMessage(input.Question)
        msg.AddImageData([]byte(input.PageImageBase64), "high", "png")
        // Appel LLM vision pour interpréter la page visuellement
        // (tables complexes, graphiques, etc.)
        req := &model.Request{
            Messages: []model.Message{msg},
            GenerationConfig: model.GenerationConfig{Stream: false},
        }
        // ... appel model.GenerateContent et retour du texte
        return "extracted text from vision", nil
    },
    function.WithName("vision_parse_page"),
    function.WithDescription("Use vision model to interpret a complex page image (tables, charts)"),
)

parseAgent := llmagent.New("page-parser",
    llmagent.WithModel(llm),
    llmagent.WithTools([]tool.Tool{parseDocTool, visionPageTool}),
    llmagent.WithInstruction(`Parse the document.
Use parse_document to extract all text (includes automatic OCR for images).
If specific pages have complex tables or charts that need visual interpretation,
use vision_parse_page to get a better extraction.
Return all parsed content organized by page.`),
    llmagent.WithGenerationConfig(model.GenerationConfig{
        Temperature: floatPtr(0.1),
        Stream:      true,
    }),
)
```

### Agent 3: Tree Builder

```go
type DocumentTree struct {
    Nodes []TreeNode `json:"nodes"`
}

type TreeNode struct {
    ID       string     `json:"id"`        // kebab-case
    Title    string     `json:"title"`
    Summary  string     `json:"summary"`   // 15-25 mots
    Text     string     `json:"text"`      // Texte brut complet
    Children []TreeNode `json:"children"`
}

treeAgent := llmagent.New("tree-builder",
    llmagent.WithModel(llm),
    llmagent.WithInstruction(`Build a hierarchical document tree from the TOC and parsed pages.
Each node must have:
- id: unique kebab-case identifier
- title: section heading
- summary: one sentence, 15-25 words
- text: full raw text of the section
- children: nested sub-sections
Max depth: 2 levels.`),
    llmagent.WithStructuredOutputJSON(
        (*DocumentTree)(nil),
        true,
        "Hierarchical document tree",
    ),
    llmagent.WithGenerationConfig(model.GenerationConfig{
        Temperature: floatPtr(0.2),
        MaxTokens:   intPtr(8000),
        Stream:      false,
    }),
)
```

### Agent 4: Summarizer

```go
type DocumentMeta struct {
    Description string   `json:"description"`   // Description du document entier
    TopSections []string `json:"top_sections"`   // Titres top-level pour le listing
}

summarizerAgent := llmagent.New("summarizer",
    llmagent.WithModel(llm),
    llmagent.WithInstruction(`Generate a description of the entire document and list the top-level section titles.
The description should be 1-2 sentences explaining what this document is about.`),
    llmagent.WithStructuredOutputJSON(
        (*DocumentMeta)(nil),
        true,
        "Document metadata with description",
    ),
    llmagent.WithGenerationConfig(model.GenerationConfig{
        Temperature: floatPtr(0.3),
        Stream:      false,
    }),
)
```

---

## Assemblage dans le service

```go
// core/service/indexer_service.go

type IndexerService struct {
    pipeline agent.Agent          // Le ChainAgent
    runner   runner.Runner
    docRepo  port.DocumentRepository
    idxRepo  port.IndexRepository
    logger   *slog.Logger
}

func NewIndexerService(
    llm model.Model,
    pdfParser port.PDFParser,
    ocrProvider port.OCRProvider,
    docRepo port.DocumentRepository,
    idxRepo port.IndexRepository,
    logger *slog.Logger,
) *IndexerService {

    // Créer les tools
    parsePagesTool := buildParsePagesTool(pdfParser)
    ocrPageTool := buildOCRPageTool(ocrProvider)

    // Créer les agents
    tocAgent := buildTOCAgent(llm)
    parseAgent := buildParseAgent(llm, parsePagesTool, ocrPageTool)
    treeAgent := buildTreeAgent(llm)
    summarizerAgent := buildSummarizerAgent(llm)

    // Assembler le pipeline
    pipeline := chainagent.New("indexing-pipeline",
        chainagent.WithSubAgents([]agent.Agent{
            tocAgent,
            parseAgent,
            treeAgent,
            summarizerAgent,
        }),
    )

    r := runner.NewRunner("logidoc-indexer", pipeline,
        runner.WithSessionService(inmemory.NewSessionService()),
    )

    return &IndexerService{
        pipeline: pipeline,
        runner:   r,
        docRepo:  docRepo,
        idxRepo:  idxRepo,
        logger:   logger,
    }
}

func (s *IndexerService) Index(ctx context.Context, docID string, content string) error {
    s.logger.Info("starting indexation", "doc_id", docID)

    // Update status
    s.docRepo.UpdateStatus(ctx, docID, domain.StatusIndexing, "")

    // Run the pipeline
    sessionID := fmt.Sprintf("index-%s", docID)
    msg := model.NewUserMessage("Index this document:\n\n" + content)

    eventChan, err := s.runner.Run(ctx, "indexer", sessionID, msg)
    if err != nil {
        s.docRepo.UpdateStatus(ctx, docID, domain.StatusError, err.Error())
        return err
    }

    // Collect result from the last agent (summarizer)
    var finalResult string
    for event := range eventChan {
        if event.Error != nil {
            s.logger.Error("pipeline error", "error", event.Error.Message)
            s.docRepo.UpdateStatus(ctx, docID, domain.StatusError, event.Error.Message)
            return fmt.Errorf("pipeline error: %s", event.Error.Message)
        }
        // Capture le dernier output complet
        if event.Response != nil && !event.Response.IsPartial {
            if len(event.Response.Choices) > 0 {
                finalResult = event.Response.Choices[0].Message.Content
            }
        }
    }

    // Parse et sauvegarde
    var tree DocumentTree
    json.Unmarshal([]byte(finalResult), &tree)

    s.idxRepo.Save(ctx, &domain.Index{DocID: docID, Tree: tree.Nodes})
    s.docRepo.UpdateStatus(ctx, docID, domain.StatusReady, "")

    s.logger.Info("indexation complete", "doc_id", docID)
    return nil
}
```

---

## Pour la retrieval (navigation dans l'arbre)

La retrieval utilise un **LLMAgent simple** (pas de chain):

```go
// core/service/retrieval_service.go

func (s *RetrievalService) Search(ctx context.Context, query domain.SearchQuery) (*domain.SearchResult, error) {
    // Pour chaque document cible:
    //   1. Charger l'index (tree compacte: titres + summaries seulement)
    //   2. Appeler le LLM pour naviguer
    //   3. Extraire le texte des nodes sélectionnés (pas de LLM)

    // Étape 2: Navigation via model.Model direct (pas besoin d'un agent complet)
    treeJSON := serializeCompactTree(index.Tree)  // titres + summaries only

    req := &model.Request{
        Messages: []model.Message{
            model.NewSystemMessage(navigationPrompt),
            model.NewUserMessage(fmt.Sprintf("TOC:\n%s\n\nQuery: %s", treeJSON, query.Query)),
        },
        GenerationConfig: model.GenerationConfig{
            Temperature: floatPtr(0.1),
            Stream:      false,
        },
    }

    respChan, _ := s.llm.GenerateContent(ctx, req)
    // Parse node IDs from response
    // Extract text from those nodes (pure Go, no LLM)
    // Return results
}
```

Ici on utilise `model.Model` directement car la navigation est un **appel simple**:
query + arbre → node IDs. Pas besoin d'orchestration.

---

## Résumé: Quand utiliser quoi

| Besoin | Composant trpc-agent-go | Pourquoi |
|--------|-------------------------|----------|
| **Pipeline indexation** | `ChainAgent` + `LLMAgent` + `FunctionTool` | Séquentiel, multi-étapes, outils custom |
| **Structured output** | `WithStructuredOutputJSON()` | JSON validé par schema à chaque étape |
| **Parsing PDF** | `knowledge/document/reader/pdf` (pdfcpu) | **Déjà dans le framework** — texte page par page |
| **OCR tables/images** | `knowledge/ocr/tesseract` | **Déjà dans le framework** — s'injecte dans le PDF reader |
| **Vision VLM (fallback)** | `model.Message.AddImageData()` | **Déjà dans le framework** — multimodal natif |
| **Navigation arbre** | `model.Model` direct | Un seul appel LLM simple |
| **One-shot (pas de chat)** | `Runner` + `inmemory.SessionService` | Session éphémère, pas de persistence |

### Composants réutilisés du framework (zéro code custom)

```
trpc-agent-go/knowledge/document/reader/pdf   → PDF parsing (pdfcpu + ledongthuc/pdf)
trpc-agent-go/knowledge/ocr/tesseract         → OCR sur images extraites du PDF
trpc-agent-go/model.Message.AddImageData()    → Vision multimodale (VLM fallback)
trpc-agent-go/model/openai                    → Provider LLM OpenAI
trpc-agent-go/model/anthropic                 → Provider LLM Anthropic
trpc-agent-go/agent/chainagent                → Pipeline séquentiel
trpc-agent-go/agent/llmagent                  → Agents avec tools + structured output
trpc-agent-go/tool/function                   → Custom tools (FunctionTool)
trpc-agent-go/runner                          → Orchestration + session
trpc-agent-go/session/inmemory                → Session éphémère (one-shot)
trpc-agent-go/tool/mcp                        → MCP client (pour les agents consommateurs)
trpc.group/trpc-go/trpc-mcp-go               → MCP server (pour exposer les tools)
```
