# PageIndex — Methodology

## Four-Phase Pipeline

```
Phase 1 (Index):    Full Document  →  LLM  →  TOC Tree (JSON)
Phase 2A (Navigate): TOC + Query   →  LLM  →  Selected Node IDs
Phase 2B (Extract):  Node IDs      →  Code →  Raw Section Text
Phase 2C (Answer):   Section Text  →  LLM  →  Final Answer + Citations
```

The document is processed ONCE during indexing. All subsequent queries only use
targeted excerpts (92-93% token reduction on average).

---

## Phase 1: Indexing — Build the Tree

**Input**: Full document (PDF or Markdown)
**Output**: JSON hierarchical tree (Table of Contents)

### Process:
1. Send full document to LLM with structured prompt
2. LLM generates a hierarchical JSON index respecting natural document boundaries
3. Retry logic (3 attempts) with JSON repair
4. Fallback to regex-based heading scanning if LLM parsing fails

### Configuration:
- `--max-pages-per-node`: Default 10
- `--max-tokens-per-node`: Default 20,000
- `--toc-check-pages`: Default 20 (auto-detect existing TOC within first N pages)
- Max tree depth: 2 levels

### LLM Prompt Constraints:
- Output only valid JSON array
- Each node must have exactly 4 required keys
- Summaries must be exactly one sentence (15-25 words)
- NodeIds must use kebab-case format

---

## Phase 2A: Reasoning-Based Navigation

**Input**: User query + TOC tree (titles and summaries only — NOT full content)
**Output**: 1-3 most relevant node IDs

### Process:
The LLM receives the query and the lightweight TOC and reasons about which sections
are most likely to contain the answer. The prompt says:

> "Read the user's query, study the TOC summaries, and identify which node IDs
> are MOST likely to contain the answer."

Output format: comma-separated node IDs, no explanation.

### Key Insight:
This is where the "reasoning" happens — the LLM uses semantic understanding of
section titles and summaries to navigate, not vector similarity scores.

---

## Phase 2B: Targeted Extraction (No LLM)

**Input**: Selected node IDs + raw document
**Output**: Extracted section text

### Process:
1. Scan raw document line-by-line (no LLM involved)
2. Match target section headings by node ID
3. Capture text until next section boundary
4. Return extracted text, or fall back to full document

This is a pure code operation — fast and deterministic.

---

## Phase 2C: Answer Synthesis

**Input**: Extracted section text + original query
**Output**: Answer with citations

### Process:
The LLM receives ONLY the small extracted text and the query with instruction:

> "Answer the user's question using ONLY the information in this excerpt."

This grounds the answer in specific document sections and enables precise citations.

---

## Five-Step Iterative Retrieval Loop

For complex queries, the retrieval is iterative:

1. **Read the Table of Contents** — understand document structure
2. **Select a Section** — choose the most relevant node
3. **Extract Relevant Information** — parse selected content
4. **Assess Sufficiency** — is more information needed?
5. **Answer the Question** — generate response once adequate context is collected

The loop can:
- Follow in-document references (e.g., "see Appendix G")
- Consider chat history for multi-turn awareness
- Fetch neighboring sections when content is incomplete
