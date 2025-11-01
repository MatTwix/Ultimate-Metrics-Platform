FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/notification-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/notification-service ./cmd/notification-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/notification-service .

COPY ./services/notification-service/configs ./configs

CMD ["/app/notification-service"]