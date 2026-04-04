# logidoc_get_sections

Get the full text of specific sections.

## Parameters

| Name | Type | Required |
|------|------|----------|
| `doc_id` | string | yes |
| `node_ids` | string | yes |

`node_ids` is comma-separated: `"chapter-1,section-1-2"`.

## Returns

JSON array of sections with full text, page numbers, and metadata.

```json
[
  {
    "ID": "section-1-1-context",
    "Title": "Context",
    "Summary": "Background and motivation.",
    "Text": "Ora is a platform designed for interacting with conversational AI agents...",
    "StartPage": 1,
    "EndPage": 3,
    "Children": []
  }
]
```

If table extraction is enabled, the text includes markdown tables. If image description is enabled, it includes `[Image: description]` annotations.

## When to use

After identifying relevant sections from the TOC or search results, retrieve their full text to answer the user's question. Always cite the section title and page numbers in your response.
