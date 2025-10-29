FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/collector-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/collector-service ./cmd/collector-service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/collector-service .

COPY ./services/collector-service/configs ./configs

EXPOSE 8080

CMD ["/app/collector-service"]