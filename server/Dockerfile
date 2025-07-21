# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make protoc protobuf-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 secretly && \
    adduser -D -s /bin/sh -u 1001 -G secretly secretly

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/secretly-server .

# Copy configuration template
COPY --from=builder /app/secretly.yaml.tpl ./secretly.yaml

# Create directories for data and certificates
RUN mkdir -p data certs keys && \
    chown -R secretly:secretly /app

# Switch to non-root user
USER secretly

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV SECRETLY_CONFIG_PATH=/app/secretly.yaml

# Run the server
CMD ["./secretly-server"]