# Build stage
FROM golang:1.25-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh -u 1001 appuser

# Set working directory
WORKDIR /build

# Copy database module (required dependency)
COPY database/ ./database/

# Copy logger module
COPY logger/ ./logger/

# Set working directory to logger
WORKDIR /build/logger

# Download dependencies (this layer will be cached if go.mod/go.sum don't change)
RUN go mod download && go mod verify

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o logger ./cmd/logger

# Check if GeoIP database is present (just for logging)
RUN if [ -f GeoLite2-Country.mmdb ]; then \
        echo "GeoLite2-Country.mmdb found - will be included"; \
    else \
        echo "GeoLite2-Country.mmdb not found - IP geolocation will be disabled"; \
    fi

# Final stage - minimal runtime image
FROM alpine:3.19

# Install ca-certificates for HTTPS and create app user
RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -s /bin/sh -u 1001 appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/logger/logger /app/logger

# Copy GeoIP database if it exists (optional - using shell test to avoid Docker errors)
RUN --mount=type=bind,from=builder,source=/build/logger,target=/build/logger \
    if [ -f /build/logger/GeoLite2-Country.mmdb ]; then \
        cp /build/logger/GeoLite2-Country.mmdb /app/GeoLite2-Country.mmdb; \
        chown appuser:appuser /app/GeoLite2-Country.mmdb; \
    fi

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
CMD ["./logger"]