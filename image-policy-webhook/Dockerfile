FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o webhook-server main.go

# Final image
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/webhook-server .
COPY tls.crt /etc/webhook/certs/tls.crt
COPY tls.key /etc/webhook/certs/tls.key

EXPOSE 8443

CMD ["./webhook-server", "--tls-cert=/etc/webhook/certs/tls.crt", "--tls-key=/etc/webhook/certs/tls.key"]
