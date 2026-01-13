# Build stage
FROM golang:1.25-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh -u 1001 appuser

# Copy go.mod and go.sum only
COPY monitor/go.mod monitor/go.sum ./monitor/

WORKDIR /build/monitor
RUN go mod download && go mod verify

# Now copy the rest of the source
COPY database/ /build/database/
COPY monitor/ /build/monitor/

# Download dependencies (this layer will be cached if go.mod/go.sum don't change)
RUN go mod download && go mod verify

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o monitor ./cmd/monitor

# Final stage - minimal runtime image
FROM alpine:3.19

# Install ca-certificates for HTTPS and create app user
RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -s /bin/sh -u 1001 appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/monitor/monitor /app/monitor

# Create logs directory and set permissions
RUN mkdir -p /app/logs && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/api/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV LOG_LEVEL=info

# Run the application
CMD ["./monitor"]