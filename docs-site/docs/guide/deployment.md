# Deployment

## Docker Compose (recommended)

```bash
cp deploy/.env.prod.example deploy/.env.prod
# Edit deploy/.env.prod with your production values
make prod
```

## Docker image

Build and run the production image:

```bash
docker build -f deploy/Dockerfile.prod -t logidoc-server .
docker run -p 7042:7042 -p 7043:7043 --env-file .env logidoc-server
```

## Binary

```bash
go build -o logidoc-server ./cmd/server
./logidoc-server
```

Requires `pdftotext` (poppler-utils) installed on the host.

## Reverse proxy

In production, put logidoc behind nginx or Caddy for TLS:

```nginx
server {
    server_name logidoc.your-domain.com;

    location / {
        proxy_pass http://localhost:7042;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        client_max_body_size 50M;
    }
}
```

## Health check

```bash
curl http://localhost:7042/health
```

Returns:

```json
{
  "status": "ok",
  "version": "1.0.0",
  "mongo": "ok"
}
```

Use this for load balancer health checks and Kubernetes probes.
