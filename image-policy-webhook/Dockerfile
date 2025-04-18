# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY main.go .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o webhook main.go

# Stage 2: Create the runtime image
FROM alpine:3.18

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/webhook .

# Create directory for TLS certificates
RUN mkdir -p /app/certs

# Copy TLS certificates (assumed to be in the same directory as Dockerfile)
# Replace with actual certificate paths or mount them at runtime
COPY tls.crt /app/certs/tls.crt
COPY tls.key /app/certs/tls.key

# Expose port 8080
EXPOSE 8080

# Run the webhook
CMD ["./webhook"]