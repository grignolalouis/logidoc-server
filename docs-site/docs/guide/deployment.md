# Deployment

## Docker Compose

```bash
cp deploy/.env.prod.example deploy/.env.prod
# Edit deploy/.env.prod
make prod
```

## Docker Image

```bash
docker build -f deploy/Dockerfile.prod -t logidoc-server .
docker run -p 7042:7042 -p 7043:7043 --env-file .env logidoc-server
```

Requires MongoDB accessible at `MONGO_URI`.

## Binary

```bash
go build -o logidoc-server ./cmd/server
./logidoc-server
```

Requires `pdftotext` (poppler-utils) on the host.

## Reverse Proxy

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

## Health Check

```
GET /health → {"status": "ok", "version": "1.0.0", "mongo": "ok"}
```
