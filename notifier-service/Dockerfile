FROM golang:1.26.1-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o notifier ./cmd/notifier

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/notifier .
CMD ["./notifier"]
