# Authentication

API key authentication is optional. When enabled, all `/v1/*` endpoints require a `Bearer` token.

## Enable

Set `API_KEY` in your `.env`:

```env
API_KEY=sk-logidoc-your-secret-key
```

## Usage

Include the key in the `Authorization` header:

```bash
curl -H "Authorization: Bearer sk-logidoc-your-secret-key" \
  http://localhost:7042/v1/documents
```

With the SDK:

```typescript
const client = new LogidocClient({
  baseUrl: "https://logidoc.your-domain.com",
  headers: { Authorization: "Bearer sk-logidoc-your-secret-key" },
});
```

## What's protected

| Route | Auth required |
|-------|--------------|
| `/v1/*` | Yes (when API_KEY is set) |
| `/health` | No |
| `/version` | No |
| `/ui/*` | No |

## Security notes

- Never expose the API key in browser-side code
- Use the SDK from your backend, not from the frontend
- For the admin UI, protect it with a reverse proxy or private network
