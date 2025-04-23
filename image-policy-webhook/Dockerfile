# ----------- Stage 1: Builder ------------
    FROM golang:1.23-alpine AS builder

    # Set working directory
    WORKDIR /app
    
    # Install OpenSSL and bash for generating certs
    RUN apk --no-cache add openssl bash
    
    # Copy all source files and cert generation script
    COPY . .
    
    # Make cert script executable and generate certs
    RUN chmod +x ./generate-certs.sh && ./generate-certs.sh
    
    # Build the Go application
    RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webhook-server .
    
    # ----------- Stage 2: Final Image ------------
    FROM alpine:3.19
    
    # Install CA certificates and required tools
    RUN apk --no-cache add ca-certificates bash && \
        update-ca-certificates
    
    # Create working directory
    WORKDIR /app
    
    # Copy the built webhook server binary
    COPY --from=builder /app/webhook-server /app/
    
    # Copy generated certs from builder
    COPY --from=builder /app/certs /app/certs
    COPY --from=builder /app/keys /app/keys
    
    # Set environment variables
    ENV CERT_FILE=/app/certs/webhook-server.crt
    ENV KEY_FILE=/app/keys/webhook-server.key
    ENV PORT=8443
    
    # Create a non-root user and change ownership
    RUN addgroup -S webhook && adduser -S webhook -G webhook && \
        chown -R webhook:webhook /app
    
    # Switch to non-root user
    USER webhook
    
    # Expose the webhook server port
    EXPOSE 8443
    
    # Start the webhook server
    ENTRYPOINT ["/app/webhook-server"]
    