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

## Parsing

Logidoc uses **pdftotext** (poppler) for PDF text extraction. It runs page by page with `-layout` to preserve spacing and alignment.

For non-PDF formats (DOCX, PPTX, HTML, EPUB, XLSX), **pandoc** converts to markdown first.

Scanned PDFs (image-only pages) are not yet supported. OCR support is planned for a future release.

## Structure detection

The indexer uses an LLM agent with two tools: `get_page_count` and `read_pages`. The agent reads the first 5 pages looking for a table of contents.

**If a TOC with page numbers exists**, the agent extracts the section list directly. This costs ~2k tokens and takes ~10 seconds. The agent does NOT read the rest of the document.

**If no TOC exists**, the indexer scans the entire document in chunks of 10 pages. Each chunk is sent to the LLM with the previously detected sections for continuity. This costs more (~30k tokens for a 100-page doc) but covers every page.

The LLM returns a **flat list** of sections:

```json
[
  {"title": "Introduction", "id": "introduction", "level": 1, "start_page": 3, "summary": "..."},
  {"title": "Context", "id": "context", "level": 2, "start_page": 3, "summary": "..."},
  {"title": "Methodology", "id": "methodology", "level": 1, "start_page": 15, "summary": "..."}
]
```

Go code then computes `end_page` for each section and builds the hierarchical tree from the `level` field. No LLM needed for tree construction.

## Page calibration

Academic PDFs often have a title page, TOC pages, and preface before "page 1" of the actual content. The TOC says "Chapter 1 ... page 1" but it's actually page 7 of the PDF.

Logidoc detects this offset automatically. It searches for section titles in the extracted text, finds the physical page where they appear, and adjusts all page numbers. A verification step then checks each section title against its assigned page and corrects any remaining misalignment.

## Table extraction

`pdftotext -layout` preserves table alignment but columns can get garbled. Logidoc has two strategies:

1. **Heuristic detection** (free, fast): scans the text for lines with column-aligned whitespace gaps, groups them, and converts to markdown tables.

2. **VLM fallback** (opt-in, costs tokens): if a page looks tabular but the heuristic found nothing, it renders the page as an image and asks a vision model to extract tables as markdown.

Enable with `INDEXER_ENABLE_TABLE_EXTRACTION=true`.

## Image description

PDFs contain embedded images (diagrams, charts, screenshots) that are invisible to text extraction. Logidoc can extract them and describe them:

1. **pdfimages** extracts embedded images from the PDF
2. Each image is sent to the vision model with `AddImageData()`
3. The description is appended to the page content as `[Image: description]`

Enable with `INDEXER_ENABLE_IMAGE_DESCRIPTION=true`.

## Two models

Logidoc supports two separate models for different tasks:

- **Main model** (`LLM_*`) — structure detection, TOC extraction, chunking, subdivision. Text-only, fast, cheap. Haiku or gpt-4o-mini recommended.
- **Vision model** (`VISION_*`) — table VLM fallback, image description. Needs vision capability. GPT-4o-mini or Sonnet recommended.

If no vision model is configured, the main model is used for everything. This works but vision-capable models tend to hallucinate less on image tasks.

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

## Token costs

| Document | With TOC | Without TOC |
|----------|----------|-------------|
| 30 pages | ~6k tokens | ~30k tokens |
| 100 pages | ~8k tokens | ~80k tokens |
| 400 pages | ~10k tokens | ~200k tokens |

Per retrieval query: ~2k tokens (TOC) + section text.

With table extraction and image description enabled, add ~500 tokens per page with tables and ~1k tokens per image.
