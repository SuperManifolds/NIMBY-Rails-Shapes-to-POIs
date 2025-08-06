# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

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

# Runtime stage - using distroless static for minimal size and fast cold starts
FROM gcr.io/distroless/static-debian12:nonroot

# Copy ca-certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/nimby_shapetopoi .

# Copy static files
COPY --from=builder /app/static ./static

# Expose port (Cloud Run will set PORT environment variable)
EXPOSE 8080

# Run the application (will read PORT environment variable automatically)
ENTRYPOINT ["./nimby_shapetopoi", "--server"]