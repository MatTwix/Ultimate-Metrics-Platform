FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/persister-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/persister-service ./services/persister-service/cmd/persister-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/persister-service .

COPY ./services/persister-service/configs ./configs
COPY ./services/persister-service/internal/database/migrations ./internal/database/migrations

CMD ["/app/persister-service"]