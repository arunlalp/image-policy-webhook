# Use a multi-stage build
FROM golang:1.23-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and if present go.sum files
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webhook-server .

# Create the final image
FROM alpine:3.19

# Add CA certificates and OpenSSL for HTTPS and certificate generation
RUN apk --no-cache add ca-certificates openssl bash && \
    update-ca-certificates

# Create directories for SSL certificates
RUN mkdir -p /app/certs /app/keys

# Copy the webhook server binary
COPY --from=builder /app/webhook-server /app/

# Copy the certificate generation script
COPY generate-certs.sh /app/
RUN chmod +x /app/generate-certs.sh
RUN ./generate-certs.sh

# Copy the generated certificates if they exist
COPY certs/ /app/certs/
COPY keys/ /app/keys/

# Set working directory
WORKDIR /app

# Configure environment variables
ENV CERT_FILE=/app/certs/webhook-server.crt
ENV KEY_FILE=/app/keys/webhook-server.key
ENV PORT=8443

# Set the entrypoint
ENTRYPOINT ["/app/webhook-server"]

# Expose port 8443
EXPOSE 8443

# Use a non-root user
RUN addgroup -S webhook && adduser -S webhook -G webhook
RUN chown -R webhook:webhook /app
USER webhook
