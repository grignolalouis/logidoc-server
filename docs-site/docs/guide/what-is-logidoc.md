# What is logidoc?

logidoc indexes documents into hierarchical trees that AI agents can browse and search.

## The problem

LLM agents need to find information in documents. Traditional approaches use vector databases and embeddings — they chunk documents into pieces, embed them, and do similarity search. This loses document structure, mixes up sections, and requires maintaining a vector DB.

## The approach

logidoc uses the **PageIndex** methodology: instead of vectors, it builds a **table of contents** with summaries for each section. An agent reads the TOC, reasons about which sections are relevant, and retrieves only what it needs.

```
PDF → Extract text page by page
    → LLM identifies sections and page ranges
    → Server builds hierarchical tree with summaries
    → Stored in MongoDB
```

When an agent queries:

```
1. get_toc(doc_id)     → titles + summaries (~2k tokens)
2. agent reasons       → picks 1-3 sections
3. get_sections(ids)   → full text of those sections
```

The agent never sees the full document. A 400-page book costs ~5k tokens to index and ~2k tokens per query.

## How it works

### Indexation

1. **Parse** — Extract text page by page using `pdftotext`
2. **Detect TOC** — An LLM agent reads the first pages looking for a table of contents
3. **Build structure** — If TOC found, map sections to page ranges directly. If not, scan the document in chunks of 10 pages
4. **Fill text** — Server fills each section with the raw text from its page range (no LLM)
5. **Store** — Save the tree in MongoDB

### Retrieval

Agents connect via MCP or the HTTP API. They see three tools:

- **list_documents** — what's available
- **get_toc** — browse the structure (titles, summaries, page numbers)
- **get_sections** — get the full text of specific sections
