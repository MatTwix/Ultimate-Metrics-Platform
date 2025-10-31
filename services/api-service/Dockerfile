FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/api-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api-service ./cmd/api-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/api-service .

COPY ./services/api-service/configs ./configs

CMD ["/app/api-service"]