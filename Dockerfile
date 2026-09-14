# Stage 1: Build binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy dependencies first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build small static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o findit-api main.go

# Stage 2: Production runtime image
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Copy binary from build stage
COPY --from=builder --chown=appuser:appgroup /app/findit-api /app/findit-api

EXPOSE 8094

ENTRYPOINT ["/app/findit-api"]
