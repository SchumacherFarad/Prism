# Build stage - using debian-based image for CGO support (required by go-sqlite3)
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Install build dependencies for CGO and SQLite
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    git \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGO enabled (required for go-sqlite3)
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /prism ./cmd/prism

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies
# - ca-certificates: for HTTPS requests
# - tzdata: timezone support
# - libsqlite3-0: SQLite runtime library
# - chromium: for Playwright/TEFAS scraping
# - nodejs npm: temporarily for installing Playwright driver
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    libsqlite3-0 \
    chromium \
    wget \
    curl \
    gnupg \
    && mkdir -p /etc/apt/keyrings \
    && curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg \
    && echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_20.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list \
    && apt-get update \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install Playwright driver (required by playwright-go)
# This is needed because playwright-go requires its own driver binaries
# We install Playwright's Chromium to its default cache location
RUN npm install -g playwright@1.52.0 && \
    npx playwright install --with-deps chromium && \
    npm uninstall -g playwright && \
    apt-get purge -y nodejs && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/* /etc/apt/sources.list.d/nodesource.list

# Copy binary from builder
COPY --from=builder /prism /app/prism

# Copy config example (actual config should be mounted)
COPY config.example.yaml /app/config.example.yaml

# Create data directory
RUN mkdir -p /app/data

# Set environment variables for Playwright
# npm installs browsers to /root/.cache/ms-playwright
# PLAYWRIGHT_BROWSERS_PATH tells playwright-go where to find the browsers
ENV PLAYWRIGHT_BROWSERS_PATH=/root/.cache/ms-playwright
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1

EXPOSE 8080

ENTRYPOINT ["/app/prism"]
