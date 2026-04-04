# Indexer Findings

## Current State

The indexer parses PDFs via `pdftotext`, detects TOC via an LLM agent, builds a hierarchical tree with page ranges, and fills text from parsed pages.

## Weaknesses Identified

| # | Issue | Impact | Current behavior |
|---|-------|--------|-----------------|
| 1 | Tables garbled by pdftotext | High | Columns merge, data unreadable |
| 2 | Images/diagrams ignored | High | Agent sees nothing for visual content |
| 3 | Only PDF + plaintext supported | High | No docx, pptx, html, epub |
| 4 | Scanned PDFs produce empty text | High | pdftotext returns nothing on image-based PDFs |
| 5 | Large sections (50+ pages) not subdivided | Medium | Huge text returned to agent |
| 6 | Summaries from titles only, not content | Medium | Generic summaries hurt retrieval accuracy |
| 7 | No cross-document search | Medium | Agent does N+1 calls to scan all docs |
| 8 | Page calibration imprecise for complex docs | Low | Offset detection fails on multi-numbering schemes |
| 9 | No incremental re-indexing | Low | Full reprocess required |
| 10 | No token cost tracking | Low | No visibility on indexation cost |

## Test Results

Tested on two documents:
- II3510_Ora_Project_Report.pdf (32 pages, with TOC)
- The.Go.Programming.Language.pdf (380 pages, with TOC)

| Metric | Ora Report | Go Book |
|--------|-----------|---------|
| Recall (with TOC) | 100% | ~95% |
| Indexation time | 16s | ~45s |
| Tokens used | 5,897 | ~8,000 |
| LLM calls | 2 | 3 |
| Pages read | 5 | 5 |

The TOC path is fast and accurate. The no-TOC path (sequential chunking) is 6x more expensive but still produces good results (90%+ recall).
