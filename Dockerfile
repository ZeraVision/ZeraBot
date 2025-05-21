# Build stage: Use official Go 1.23.2 Alpine image
FROM golang:1.23.2-alpine AS builder

# Install Git for fetching Go module dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -o /app/bin/zerabot

# Final stage
FROM debian:bookworm-slim

# Install required packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user and required directories
RUN groupadd -r zerabot && \
    useradd -r -g zerabot -d /app -s /sbin/nologin -c "ZeraBot user" zerabot && \
    mkdir -p /app/certs && \
    chown -R zerabot:zerabot /app && \
    chmod -R 755 /app && \
    chmod 700 /app/certs

WORKDIR /app

# Copy Go binary and set permissions
COPY --from=builder /app/bin/zerabot /app/zerabot
RUN chown -R zerabot:zerabot /app && \
    chmod +x /app/zerabot

# Switch to non-root user
USER zerabot

# Expose the port the app runs on
EXPOSE 8080 443 50051

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the app
ENTRYPOINT ["/app/zerabot"]
