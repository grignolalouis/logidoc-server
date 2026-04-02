# PageIndex vs Vector RAG — Comparison

## Side-by-Side

| Dimension             | Vector RAG                          | PageIndex (Reasoning RAG)              |
|-----------------------|-------------------------------------|----------------------------------------|
| **Retrieval method**  | Embedding similarity search         | LLM reasoning over tree structure      |
| **Index storage**     | External vector database            | In-context JSON (within LLM context)   |
| **Chunking**          | Fixed-size fragments (lossy)        | Natural document sections (lossless)   |
| **Query matching**    | Surface-level similarity            | Intent-driven inference                |
| **Cross-references**  | Fails without preprocessing         | Follows TOC structure naturally         |
| **Multi-turn**        | Isolated queries                    | Chat history aware                     |
| **Explainability**    | Opaque similarity scores            | Exact node IDs and section citations   |
| **Latency**           | Fast (single vector lookup)         | Slower (multiple LLM calls)            |
| **Best for**          | Large collections, semantic search  | Structured docs, precise QA            |
| **Scaling**           | Scales to millions of docs easily   | Best for single/few docs at a time     |

## Where PageIndex Wins

1. **Structured documents** — financial reports, legal filings, technical manuals
2. **Precise QA** — "What is the debt-to-equity ratio in Q3?" needs exact section, not similar chunk
3. **Cross-reference following** — "see Appendix G" type references
4. **Context preservation** — no information loss from chunking
5. **Auditability** — every answer traces back to specific sections

## Where Vector RAG Wins

1. **Large document collections** — searching across thousands of documents
2. **Semantic discovery** — finding topically related but structurally unrelated content
3. **Speed** — single vector lookup vs. multiple LLM reasoning steps
4. **Cost** — no LLM calls for retrieval phase

## Key Insight

They are complementary, not competing:
- Use **vector search** to find WHICH documents are relevant (collection-level)
- Use **PageIndex** to find precise answers WITHIN those documents (document-level)
