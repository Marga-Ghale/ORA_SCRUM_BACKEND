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
RUN apk --no-cache add ca-certificates postgresql-client curl bash

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

# Entrypoint script with force migration
RUN echo '#!/bin/sh' > /entrypoint.sh && \
    echo 'set -e' >> /entrypoint.sh && \
    echo 'echo "🔄 Running database migrations..."' >> /entrypoint.sh && \
    echo '# Force to version 3 to clear any dirty state' >> /entrypoint.sh && \
    echo 'migrate -path ./internal/db/migrations -database "$DATABASE_URL" force 3 2>/dev/null || echo "Force not needed, continuing..."' >> /entrypoint.sh && \
    echo '# Run migrations up' >> /entrypoint.sh && \
    echo 'migrate -path ./internal/db/migrations -database "$DATABASE_URL" up' >> /entrypoint.sh && \
    echo 'echo "✅ Migrations completed"' >> /entrypoint.sh && \
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