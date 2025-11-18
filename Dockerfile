# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata curl

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install templ for template generation
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy source code
COPY . .

# Generate template files
RUN templ generate

# Build the application with optimizations for smaller binary and faster startup
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -extldflags '-static'" \
    -a -installsuffix cgo \
    -trimpath \
    -o nimby_shapetopoi ./cmd/nimby_shapetopoi

# Runtime stage - using base debian for curl support
FROM debian:12-slim

# Install curl for health checks
RUN apt-get update && apt-get install -y curl ca-certificates && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r appuser && useradd -r -g appuser appuser

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/nimby_shapetopoi .

# Copy static files
COPY --from=builder /app/static ./static

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port (ECS will set PORT environment variable)
EXPOSE 8080

# Run the application (will read PORT environment variable automatically)
ENTRYPOINT ["./nimby_shapetopoi", "--server"]