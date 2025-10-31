FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/analytics-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/analytics-service ./cmd/analytics-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/analytics-service .

COPY ./services/analytics-service/configs ./configs

CMD ["/app/analytics-service"]