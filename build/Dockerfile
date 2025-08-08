# ----- Stage 1: Build binary -----
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first (for better layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimizations (static)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app/bin/appsechub ./cmd/api

# ----- Stage 2: Minimal runtime -----
# Distroless static (nonroot) for tiny, secure image
FROM gcr.io/distroless/static:nonroot

## Copy CA certs for HTTPS (optional but recommended)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

## Copy timezone data (optional)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

## Copy binary and migrations
COPY --from=builder /app/bin/appsechub /app/appsechub
COPY --from=builder /app/migrations /app/migrations

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Run the binary
CMD ["/app/appsechub"]