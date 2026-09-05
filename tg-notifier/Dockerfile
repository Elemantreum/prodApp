FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY tg-notifier/ ./tg-notifier/
COPY proto/ ./proto/

# Создаем workspace только из тех модулей,
# которые реально нужны tg-notifier
RUN go work init ./tg-notifier ./proto

WORKDIR /app/tg-notifier

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /app/server \
    ./cmd/tg-notifier


FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server .

EXPOSE 50051

CMD ["./server"]