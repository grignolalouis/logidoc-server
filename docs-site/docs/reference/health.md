# Health

Check server status and dependencies.

## Health check

```
GET /health
```

**Response** `200 OK`

```json
{
  "status": "ok",
  "version": "1.0.0",
  "mongo": "ok"
}
```

| `status` | Meaning |
|----------|---------|
| `ok` | All systems operational |
| `degraded` | MongoDB unreachable |

Use for load balancer health checks and Kubernetes liveness probes.

## Version

```
GET /version
```

```json
{
  "version": "1.0.0"
}
```

## Error format

All API errors follow this shape:

```json
{
  "error": "short description",
  "message": "detailed error message"
}
```

| Code | Meaning |
|------|---------|
| `400` | Invalid request |
| `401` | Missing or invalid API key |
| `404` | Not found |
| `409` | Document not yet indexed |
| `429` | Rate limit exceeded |
| `500` | Internal error |
