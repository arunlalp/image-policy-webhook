FROM golang:1.24-alpine as builder
WORKDIR /app
COPY main.go .
RUN go mod init webhook && go mod tidy && go build -o webhook-server

FROM alpine
RUN mkdir /certs
COPY --from=builder /app/webhook-server /webhook-server
COPY certs /certs
CMD ["/webhook-server"]
