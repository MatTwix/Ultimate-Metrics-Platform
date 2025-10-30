FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/cache-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/cache-service ./cmd/cache-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/cache-service .

COPY ./services/cache-service/configs ./configs

CMD ["/app/cache-service"]