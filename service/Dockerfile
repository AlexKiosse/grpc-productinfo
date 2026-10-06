FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY db/ ./db/
COPY service/go.mod service/go.sum ./service/
WORKDIR /app/service
RUN go mod download

COPY service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/service/server .
COPY certs/ ./certs/

ENV CERT_DIR=/root/certs

EXPOSE 50051

CMD ["./server"]
