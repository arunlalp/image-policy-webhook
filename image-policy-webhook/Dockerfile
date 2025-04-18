# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Download dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o webhook-server .

# Final stage
FROM alpine:3.18

# Install CA certificates and create app user
RUN apk --no-cache add ca-certificates && \
    adduser -D -g '' appuser

WORKDIR /app

# Copy binary and certs directory structure from builder
COPY --from=builder /app/webhook-server .
COPY --chown=appuser:appuser certs /etc/webhook/certs

# Use non-root user
USER appuser

EXPOSE 8443

ENTRYPOINT ["./webhook-server"]