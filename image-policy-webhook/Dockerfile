# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod init webhook && \
    go mod tidy && \
    go build -o webhook-server main.go

# Stage 2: Create a minimal image
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/webhook-server .

CMD ["/app/webhook-server"]
