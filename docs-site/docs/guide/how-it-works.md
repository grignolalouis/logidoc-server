# How It Works

## The problem

AI agents need to find information in documents. Traditional approaches use vector databases — they split documents into chunks, embed them, and do similarity search. This loses document structure, mixes sections, and requires maintaining an embedding pipeline.

## The approach

Logidoc builds a **table of contents** with summaries for each section. An agent reads the TOC, reasons about which sections are relevant, and retrieves only what it needs. No vectors, no embeddings.

## Indexation pipeline

When you upload and index a document, this happens:

```
Upload (stored in MongoDB)
  │
  ▼
Parse: extract text page by page
  │  PDF → pdftotext (poppler)
  │  DOCX/PPTX/HTML → pandoc → markdown
  │
  ▼
Detect structure
  │  ┌─ TOC with page numbers? → extract directly, stop
  │  ├─ TOC without page numbers? → search for titles in text
  │  └─ No TOC? → scan in chunks of 10 pages
  │
  ▼
Calibrate page numbers
  │  Logical pages (from TOC) → physical PDF pages
  │  Detects title page / preface offset automatically
  │
  ▼
Build tree + subdivide large sections
  │  Flat sections → hierarchical tree by level
  │  Sections > 20 pages → LLM splits into sub-sections
  │
  ▼
Enrich content (optional)
  │  Tables → heuristic detection + VLM fallback
  │  Images → pdfimages extraction + VLM description
  │
  ▼
Fill text from pages → save to MongoDB
```

## Two models

Logidoc supports two separate models for different tasks:

- **Main model** (`LLM_*`) — structure detection, TOC extraction, chunking, subdivision. Text-only, fast, cheap.
- **Vision model** (`VISION_*`) — table VLM fallback, image description. Needs vision capability.

If no vision model is configured, the main model is used for everything.

```env
LLM_PROVIDER=anthropic
LLM_MODEL=claude-haiku-4-5-20251001

VISION_PROVIDER=openai
VISION_MODEL=gpt-4o-mini
```

## Retrieval

Agents never see the full document. They interact through 4 tools:

1. **list_documents** → see what's available
2. **search** → keyword search across all TOCs (returns matching sections)
3. **get_toc** → browse the structure (titles + summaries + page numbers, no text)
4. **get_sections** → get full text of specific sections by ID

A typical flow:

```
Agent: search("SSE streaming")
  → 3 matching sections across 2 documents

Agent: get_sections(doc_id, "sse-streaming,sse-implementation")
  → full text of those 2 sections

Agent: answers the user's question with citations
```
