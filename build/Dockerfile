# ----- Stage 1: Build binary -----
FROM golang:1.24.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first (for better layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o /app/bin/appsechub ./cmd/api

# ----- Stage 2: Minimal runtime -----
FROM scratch

# Copy ca-certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary from builder
COPY --from=builder /app/bin/appsechub /app/appsechub

# Create non-root user (using numeric ID for better security)
USER 1000:1000

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/appsechub", "health"] || exit 1

# Run the binary
CMD ["/app/appsechub"]