# LocalMemory Dockerfile
# Multi-stage build for smaller final image

# === Go Build Stage ===
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/localmemory ./cmd/cli
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/localmemory-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/localmemory-mcp ./cmd/mcp

# === Python AI Service Stage ===
FROM python:3.11-slim AS ai-service

WORKDIR /app

# Copy Python requirements
COPY python/requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy Python code
COPY python/ .

# Default command for AI service
CMD ["python", "server.py"]

# === Final Stage ===
FROM alpine:3.19 AS final

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create data directory
RUN mkdir -p /data

# Copy binaries from builder
COPY --from=builder /build/localmemory /usr/local/bin/
COPY --from=builder /build/localmemory-server /usr/local/bin/
COPY --from=builder /build/localmemory-mcp /usr/local/bin/

# Copy default config
COPY config.json /etc/localmemory/config.json

# Set environment variables
ENV CONFIG_PATH=/etc/localmemory/config.json
ENV DATA_PATH=/data

# Expose ports
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
CMD ["localmemory-server"]
