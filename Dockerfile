# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with security and size optimizations
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w \
    -X github.com/signalridge/clinvoker/internal/app.version=${VERSION} \
    -X github.com/signalridge/clinvoker/internal/app.commit=${COMMIT} \
    -X github.com/signalridge/clinvoker/internal/app.date=${DATE}" \
    -o clinvk ./cmd/clinvk

# Runtime stage using distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12:nonroot

# Copy binary from builder
COPY --from=builder /app/clinvk /clinvk

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Use non-root user (distroless default)
USER nonroot:nonroot

# Default command
ENTRYPOINT ["/clinvk"]
CMD ["--help"]

# Metadata labels
LABEL org.opencontainers.image.title="clinvk"
LABEL org.opencontainers.image.description="Unified AI CLI wrapper for multiple backends"
LABEL org.opencontainers.image.source="https://github.com/signalridge/clinvoker"
LABEL org.opencontainers.image.vendor="SignalRidge"
LABEL org.opencontainers.image.licenses="MIT"
