# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /prism ./cmd/prism

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
# Playwright requires chromium for TEFAS scraping
RUN apk add --no-cache ca-certificates tzdata chromium

# Copy binary from builder
COPY --from=builder /prism /app/prism

# Copy config example (actual config should be mounted)
COPY config.example.yaml /app/config.example.yaml

# Create data directory
RUN mkdir -p /app/data

# Set environment variables for Playwright
ENV PLAYWRIGHT_BROWSERS_PATH=/usr/bin
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1

EXPOSE 8080

ENTRYPOINT ["/app/prism"]
