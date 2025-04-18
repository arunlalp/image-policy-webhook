FROM golang:1.21 as builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o server

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/server .
COPY tls.crt tls.key ./

EXPOSE 8443

CMD ["./server"]
