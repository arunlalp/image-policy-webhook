FROM golang:1.20 as builder

WORKDIR /app
COPY . .
RUN go mod tidy && go build -o image-policy-webhook

FROM alpine:3.18
COPY --from=builder /app/image-policy-webhook /usr/bin/
COPY tls/ /tls/
EXPOSE 8443

ENTRYPOINT ["/usr/bin/image-policy-webhook"]
