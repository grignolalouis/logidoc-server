package index

// TOCAgentInstruction is the system prompt for the agent that detects and extracts the TOC.
const TOCAgentInstruction = `You are a document structure detector. Read the first pages to find a Table of Contents.

## Tools
- get_page_count: total pages (N)
- read_pages: read pages (1-indexed, inclusive, max 10)

## Strategy
1. Call get_page_count, then read_pages(1, 5).
2. If a TOC WITH page numbers exists, extract ALL sections as a flat JSON list.
   - DO NOT read more pages. Output immediately.
3. If a TOC WITHOUT page numbers exists, read a few more pages to locate boundaries, then output.
4. If NO TOC found, output exactly: NO_TOC

## Output (if TOC found)
A JSON array of sections (FLAT, not nested). The server will build the tree.
[{"title": "Chapter 1", "id": "chapter-1", "summary": "15-25 words", "level": 1, "start_page": 3},
 {"title": "1.1 Overview", "id": "overview", "summary": "15-25 words", "level": 2, "start_page": 3},
 {"title": "Chapter 2", "id": "chapter-2", "summary": "15-25 words", "level": 1, "start_page": 15}]

## Rules
- level: 1=chapter, 2=section, 3=subsection
- start_page: the page where this section begins
- id: unique kebab-case
- summary: one sentence, 15-25 words
- Include ALL sections from page 1 to page N
- Output ONLY the JSON array or "NO_TOC"`

// ChunkInitPrompt is the prompt for the first chunk in no-TOC mode.
const ChunkInitPrompt = `Extract the document structure from this text. Return a JSON array of sections found.
Each section: {"title": "Section Title", "id": "kebab-id", "summary": "15-25 words", "level": 1, "start_page": N}
level: 1=chapter, 2=section, 3=subsection. start_page is the page number where this section starts.
Look for headings, numbered sections (1., 1.1, 2.), bold titles, or structural markers.
Output ONLY a JSON array.

Pages %d-%d of %d:
%s`

// ChunkContinuePrompt is the prompt for subsequent chunks in no-TOC mode.
const ChunkContinuePrompt = `Continue extracting document structure. Here are the sections found so far:
%s

Now extract sections from the next pages. Continue the numbering and structure.
Output ONLY a JSON array of NEW sections (not the ones above).

Pages %d-%d of %d:
%s`

const SubdividePrompt = `Extract sub-sections from this text. Return a JSON array.
Each entry: {"title": "Sub-section Title", "id": "kebab-id", "summary": "15-25 words", "level": %d, "start_page": N}
Look for headings, numbered sections, or structural markers.
Output ONLY a JSON array.

Pages %d-%d:
%s`

const ExtractTablePrompt = "Extract ALL tables from this page as markdown tables. If there are no tables, respond with 'NO_TABLES'. Output only the markdown tables, nothing else."

const DescribeImagePrompt = `Describe ONLY what you can see in this image. Do not guess, interpret, or assume anything that is not visible.
State the type of visual (diagram, screenshot, photo, chart, etc.), list any visible text labels, and describe the layout.
If you cannot identify something clearly, say so. Keep it to 2-3 sentences.`
