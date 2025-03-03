FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -o manga-reader ./cmd/server/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata sqlite

RUN adduser -D -u 1000 appuser

RUN mkdir -p /app/data /app/uploads && \
    chown -R appuser:appuser /app

WORKDIR /app

COPY --from=builder /app/manga-reader /app/

COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/.env.example /app/.env

RUN chmod +x /app/manga-reader

USER appuser


ENV SERVER_ADDRESS=:8080
ENV DB_TYPE=sqlite
ENV DB_SOURCE=/app/data/manga.db
ENV REDIS_ADDR=redis:6379

EXPOSE 8080

CMD ["/app/manga-reader"]