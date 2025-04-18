# Stage 1 - Build the binary
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o webhook-server main.go

# Stage 2 - Minimal runtime image
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/webhook-server .

CMD ["/app/webhook-server", "--tls-cert=/etc/webhook/certs/tls.crt", "--tls-key=/etc/webhook/certs/tls.key"]
