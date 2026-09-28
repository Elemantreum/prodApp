FROM golang:1.26.1-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o kafka-init ./cmd/kafka-init

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/kafka-init .
CMD ["./kafka-init"]
