FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o cloudflare-dns-rotator .

# Final minimal image
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /app/cloudflare-dns-rotator /cloudflare-dns-rotator

# Default config path
VOLUME ["/config"]

ENTRYPOINT ["/cloudflare-dns-rotator"]
CMD ["-config", "/config/config.json"]
