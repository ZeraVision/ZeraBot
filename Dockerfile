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

# Final stage: Minimal image
FROM alpine:3.19

# Install CA certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S zerabot && adduser -S zerabot -G zerabot -h /app

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bin/zerabot /app/zerabot

# Set ownership and execution permissions
RUN chown -R zerabot:zerabot /app && chmod +x /app/zerabot

# Use non-root user
USER zerabot

# Expose application ports
EXPOSE 8080 443 50051

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the app
ENTRYPOINT ["/app/zerabot"]
