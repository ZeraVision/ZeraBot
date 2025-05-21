# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .


# Build the application
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/zerabot

# Final stage
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/bin/zerabot /app/zerabot

# Copy necessary files
COPY --from=builder /app/.env /app/.env

# Create a non-root user
RUN addgroup -S zerabot && \
    adduser -S -G zerabot zerabot && \
    chown -R zerabot:zerabot /app

USER zerabot

# Expose the port the app runs on
EXPOSE 8080 443

# Command to run the executable
ENTRYPOINT ["/app/zerabot"]
