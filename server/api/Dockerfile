# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /build

# Copy module files first for dependency caching
COPY database/go.mod database/go.sum ./database/
COPY api/go.mod api/go.sum ./api/

WORKDIR /build/api
RUN go mod download && go mod verify

# Copy source
WORKDIR /build
COPY database/ ./database/
COPY api/ ./api/

WORKDIR /build/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags='-w -s -extldflags "-static"' \
    -o api ./cmd/api

# Final stage
FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata curl && \
    adduser -D -s /bin/sh -u 1001 appuser

WORKDIR /app

COPY --from=builder /build/api/api /app/api

RUN mkdir -p /app/logs && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:3000/api/health || exit 1

ENV GIN_MODE=release
ENV LOG_LEVEL=info

CMD ["./api"]
