FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init webhook && go mod tidy && go build -o webhook-server

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/webhook-server .
COPY certs /etc/webhook/certs
CMD ["/app/webhook-server"]
