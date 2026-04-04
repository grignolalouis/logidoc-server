# PageIndex — Overview

**PageIndex** is a vectorless, reasoning-based RAG framework by VectifyAI that replaces
traditional chunk-and-embed pipelines with hierarchical document indexing + LLM reasoning.

- **No Vector DB** — no embeddings, no similarity search
- **No Chunking** — documents are organized into natural sections, not arbitrary chunks
- **Explainable** — every retrieval cites exact node IDs and section titles
- **98.7% accuracy** on FinanceBench (state-of-the-art for document QA)

## Core Idea

Instead of treating retrieval as a **search problem** (find the most similar chunk),
PageIndex treats it as a **navigation problem** (reason about where the answer lives
in the document structure, like a human expert would).

## Sources

- https://github.com/VectifyAI/PageIndex
- https://pageindex.ai/blog/pageindex-intro
- https://docs.pageindex.ai/
- https://yuv.ai/blog/pageindex
- https://medium.com/@visrow/how-pageindex-works-a-step-by-step-technical-walkthrough-fca85c46a394
