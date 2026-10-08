FROM golang:1.26.1-alpine AS builder
WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN swag init -g cmd/api/main.go -o docs
RUN go build -o server ./cmd/api

FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
