FROM golang:1.25-alpine

RUN apk add --no-cache poppler-utils
RUN go install github.com/air-verse/air@v1.61.7

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

EXPOSE 7042 7043

CMD ["air", "-c", ".air.toml"]
