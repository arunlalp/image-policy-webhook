# ---------- Stage 1: Builder ----------
    FROM golang:1.23-alpine AS builder

    WORKDIR /app
    
    # Install dependencies
    RUN apk --no-cache add openssl bash
    
    # Copy all Go source files and the cert script
    COPY . .
    
    # Initialize and tidy go.mod during build
    RUN go mod init webhook-server && go mod tidy
    
    # Generate certs
    RUN chmod +x ./generate-certs.sh && ./generate-certs.sh
    
    # Build the Go app
    RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webhook-server .
    
    # ---------- Stage 2: Final Image ----------
    FROM alpine:3.19
    
    # Install required packages
    RUN apk --no-cache add ca-certificates bash && \
        update-ca-certificates
    
    WORKDIR /app
    
    # Copy the webhook binary and certs from the builder
    COPY --from=builder /app/webhook-server /app/
    COPY --from=builder /app/certs /app/certs
    
    # Environment variables (not sensitive, just paths)
    ENV CERT_FILE=/app/certs/webhook-server.crt
    ENV KEY_FILE=/app/certs/webhook-server.key
    ENV PORT=8443
    
    # Create and use non-root user
    RUN addgroup -S webhook && adduser -S webhook -G webhook && \
        chown -R webhook:webhook /app
    USER webhook
    
    # Expose the server port
    EXPOSE 8443
    
    # Entrypoint to run the server
    ENTRYPOINT ["/app/webhook-server"]
    