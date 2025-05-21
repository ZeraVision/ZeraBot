# Build stage
FROM golang:1.23.2-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -o /app/bin/zerabot

# Final stage
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies and create non-root user in a single layer
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S zerabot && \
    adduser -S -G zerabot zerabot && \
    chown -R zerabot:zerabot /app

# Copy the binary from builder
COPY --from=builder /app/bin/zerabot /app/zerabot

# Switch to non-root user
USER zerabot

# Expose the port the app runs on
EXPOSE 8080 443

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Command to run the executable
ENTRYPOINT ["/app/zerabot"]
