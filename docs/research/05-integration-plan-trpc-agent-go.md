# PageIndex Integration into trpc-agent-go — Design Notes

## What PageIndex Needs (mapped to framework concepts)

### 1. Indexing Pipeline → Could be a **Tool** or standalone **Service**
- Takes a document (PDF/Markdown) as input
- Calls LLM to generate hierarchical JSON tree index
- Stores the index for later retrieval
- This is a one-time-per-document operation

### 2. Tree Navigation → An **Agent Tool**
- Given a query + a document's TOC tree
- LLM reasons over the tree to select relevant node IDs
- Returns 1-3 node IDs

### 3. Section Extraction → A **Tool** (no LLM needed)
- Given node IDs + raw document
- Extracts raw text for those sections
- Pure code operation, fast and deterministic

### 4. Answer Synthesis → Already handled by **LLMAgent**
- The extracted text becomes context for the agent's response

## Possible Architecture in trpc-agent-go

```
┌─────────────────────────────────────┐
│          PageIndex Service          │
│  (implements a Service interface)   │
│                                     │
│  - BuildIndex(doc) → Tree JSON      │
│  - GetIndex(docID) → Tree JSON      │
│  - ExtractSections(docID, nodeIDs)  │
│    → raw text                       │
└─────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│        PageIndex Tools              │
│  (implements tool.Tool)             │
│                                     │
│  - pageindex_build: index a doc     │
│  - pageindex_search: navigate tree  │
│    + extract sections               │
│  - pageindex_list: list indexed     │
│    documents                        │
└─────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│        LLMAgent                     │
│  Uses PageIndex tools for           │
│  document understanding             │
└─────────────────────────────────────┘
```

## Alternative: As a Knowledge Implementation

The `knowledge.Knowledge` interface in trpc-agent-go:
```go
type Knowledge interface {
    Search(ctx context.Context, req *SearchRequest) (*SearchResult, error)
}
```

PageIndex could implement this interface as a reasoning-based alternative
to the existing vector-based knowledge backends. This would allow drop-in
replacement in any agent that uses knowledge retrieval.

## Key Components to Build

1. **Tree data structures** (Go structs for TOCNode, Index)
2. **Index builder** (uses LLM to generate tree from document)
3. **Tree navigator** (uses LLM to select relevant nodes given query)
4. **Section extractor** (pure Go, no LLM — maps node IDs to text)
5. **Storage backend** (persist indices — filesystem, or pluggable)
6. **Agent tools** (expose capabilities to LLMAgent)
7. **Knowledge adapter** (optional — implement knowledge.Knowledge interface)

## LLM Prompts Needed

1. **Indexing prompt**: "Given this document, generate a hierarchical JSON TOC..."
2. **Navigation prompt**: "Given this TOC and query, select 1-3 most relevant node IDs..."
3. **Answer prompt**: Already handled by the agent's own system prompt

## Considerations

- Index storage: filesystem (JSON files) is simplest, but could support any KV store
- LLM provider: should use the framework's model.Model interface (provider agnostic)
- Document parsing: PDF → text extraction needed (could leverage codeexecutor or external lib)
- Max tree depth of 2 keeps context window usage reasonable
- Multi-document search: could combine with existing vector knowledge for collection-level search
