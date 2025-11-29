# Stage 1: Build backend
FROM golang:1.23-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /linuxdo-relay \
    ./cmd/server

# Stage 2: Build frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# Copy package files
COPY web/package*.json ./
RUN npm ci

# Copy source code
COPY web/ ./

# Build frontend
RUN npm run build

# Stage 3: Final image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create app user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

WORKDIR /app

# Copy binary from backend builder
COPY --from=backend-builder /linuxdo-relay /app/linuxdo-relay

# Copy frontend from frontend builder
COPY --from=frontend-builder /app/dist /app/web/dist

# Change ownership
RUN chown -R app:app /app

# Run as non-root user
USER app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

# Run
ENTRYPOINT ["/app/linuxdo-relay"]
