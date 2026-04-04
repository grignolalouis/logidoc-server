# Search

Search across all indexed documents.

## Endpoint

```
GET /v1/search?q={query}
```

| Parameter | In | Required |
|-----------|-----|----------|
| `q` | query | yes |

## Response `200 OK`

```json
{
  "results": [
    {
      "DocID": "5d04eeed-...",
      "DocName": "report.pdf",
      "NodeID": "sse-streaming",
      "NodeTitle": "SSE Streaming",
      "Summary": "Server-sent events streaming overview.",
      "StartPage": 18,
      "EndPage": 19,
      "Score": 0.85
    }
  ]
}
```

Returns up to 10 results ranked by relevance. Keyword matching on section titles and summaries. No LLM cost.

## Errors

| Code | Reason |
|------|--------|
| `400` | Missing `q` parameter |

## SDK

::: code-group

```typescript [TypeScript]
// Not yet available in generated SDK — use fetch
const res = await fetch("http://localhost:7042/v1/search?q=SSE+streaming");
```

```go [Go]
// Not yet available in generated SDK — use HTTP client
```

```bash [cURL]
curl "http://localhost:7042/v1/search?q=SSE+streaming"
```

:::
