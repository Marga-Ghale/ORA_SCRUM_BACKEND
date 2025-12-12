# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Get prisma and generate client
RUN go get github.com/steebchen/prisma-client-go && \
    go run github.com/steebchen/prisma-client-go generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -g '' appuser

COPY --from=builder /app/main .
COPY --from=builder /app/prisma ./prisma

USER appuser
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]