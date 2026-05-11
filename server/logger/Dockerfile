# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /build

# Copy module files first for dependency caching
COPY database/go.mod database/go.sum ./database/
COPY logger/go.mod logger/go.sum ./logger/

WORKDIR /build/logger
RUN go mod download && go mod verify

# Copy source
WORKDIR /build
COPY database/ ./database/
COPY logger/ ./logger/

WORKDIR /build/logger
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags='-w -s -extldflags "-static"' \
    -o logger ./cmd/logger

# Final stage
FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -s /bin/sh -u 1001 appuser

WORKDIR /app

COPY --from=builder /build/logger/logger /app/logger

# Copy GeoIP database if present (optional - enables IP geolocation)
RUN --mount=type=bind,from=builder,source=/build/logger,target=/build/logger \
    if [ -f /build/logger/GeoLite2-Country.mmdb ]; then \
        cp /build/logger/GeoLite2-Country.mmdb /app/GeoLite2-Country.mmdb; \
        chown appuser:appuser /app/GeoLite2-Country.mmdb; \
    fi

RUN mkdir -p /app/logs && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/api/health || exit 1

ENV GIN_MODE=release
ENV LOG_LEVEL=info

CMD ["./logger"]
