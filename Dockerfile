# Build stage
FROM golang:1.24.3-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o pick6 .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /build/pick6 /app/pick6

# Copy static files
COPY --from=builder /build/static /app/static

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/pick6"]
