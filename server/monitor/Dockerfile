# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /build

COPY database/go.mod database/go.sum ./database/
COPY monitor/go.mod monitor/go.sum ./monitor/

WORKDIR /build/monitor
RUN go mod download && go mod verify

WORKDIR /build
COPY database/ ./database/
COPY monitor/ ./monitor/

WORKDIR /build/monitor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o monitor ./cmd/monitor

# Final stage
FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -s /bin/sh -u 1001 appuser

WORKDIR /app

COPY --from=builder /build/monitor/monitor /app/monitor

RUN mkdir -p /app/logs && \
    chown -R appuser:appuser /app

USER appuser

CMD ["./monitor"]
