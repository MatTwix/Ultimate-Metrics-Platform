FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/collector-service

RUN go mod download

RUN go build -o /app/server ./cmd/collector-service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .

COPY ./services/collector-service/configs ./configs
COPY ./services/collector-service/internal/database/migrations ./internal/database/migrations

EXPOSE 8080

CMD ["/app/server"]