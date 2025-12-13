# Dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build from cmd/api directory
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /app/main .
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

CMD ["./main"]