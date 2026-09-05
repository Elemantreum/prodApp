FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY backend/ ./backend/
COPY proto/ ./proto/

RUN go work init ./backend ./proto

WORKDIR /app/backend

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /app/server \
    ./cmd/api


FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]