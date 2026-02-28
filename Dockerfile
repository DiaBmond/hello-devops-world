# ==========================================
# Stage 1: Builder
# ==========================================
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git
RUN apk add --no-cache git

# Copy go mod files first (Docker cache trick)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/app

# ==========================================
# Stage 2: Production
# ==========================================
FROM alpine:latest

WORKDIR /app

# Install required packages
RUN apk add --no-cache ca-certificates dumb-init

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary and set ownership in a single layer!
COPY --chown=appuser:appgroup --from=builder /app/server .

# Switch user
USER appuser

# Expose app port
EXPOSE 8080

# Use dumb-init as entrypoint
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

# Run application
CMD ["./server"]