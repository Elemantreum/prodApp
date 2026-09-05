FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY vk-notifier/ ./vk-notifier/
COPY proto/ ./proto/

RUN go work init ./vk-notifier ./proto

WORKDIR /app/vk-notifier

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /app/server \
    ./cmd/vk-notifier


FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server .

EXPOSE 50052

CMD ["./server"]