# =========================
# Build stage
# =========================
FROM golang:1.23-alpine AS builder

# Install dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy all source code
COPY . .

# Build the Go binary (point to cmd/api where main.go is)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# =========================
# Runtime stage
# =========================
FROM alpine:latest

# Install ca-certificates and postgresql-client
RUN apk --no-cache add ca-certificates postgresql-client curl bash wget

# Install golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Set working directory
WORKDIR /root/

# Copy built binary from builder
COPY --from=builder /app/main .

# Copy migrations
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations

# Entrypoint script with better error handling
RUN echo '#!/bin/sh' > /entrypoint.sh && \
    echo 'set -e' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo 'echo "🔄 Running database migrations..."' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo '# Check if migrations directory exists and has files' >> /entrypoint.sh && \
    echo 'if [ ! -d "./internal/db/migrations" ] || [ -z "$(ls -A ./internal/db/migrations)" ]; then' >> /entrypoint.sh && \
    echo '  echo "⚠️  No migration files found, skipping migrations"' >> /entrypoint.sh && \
    echo 'else' >> /entrypoint.sh && \
    echo '  # Get current migration version' >> /entrypoint.sh && \
    echo '  CURRENT_VERSION=$(migrate -path ./internal/db/migrations -database "$DATABASE_URL" version 2>&1 || echo "none")' >> /entrypoint.sh && \
    echo '  echo "Current migration version: $CURRENT_VERSION"' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo '  # Run migrations' >> /entrypoint.sh && \
    echo '  if migrate -path ./internal/db/migrations -database "$DATABASE_URL" up; then' >> /entrypoint.sh && \
    echo '    echo "✅ Migrations completed successfully"' >> /entrypoint.sh && \
    echo '  else' >> /entrypoint.sh && \
    echo '    echo "❌ Migration failed, but continuing to start application..."' >> /entrypoint.sh && \
    echo '    echo "Check if migrations were already applied or if there is a dirty state"' >> /entrypoint.sh && \
    echo '  fi' >> /entrypoint.sh && \
    echo 'fi' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo 'echo "🚀 Starting application..."' >> /entrypoint.sh && \
    echo 'exec ./main' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

# Expose application port
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Use entrypoint script
ENTRYPOINT ["/entrypoint.sh"]