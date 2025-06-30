FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o url-shortener ./cmd/url-shortener

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/url-shortener .

ENV CONFIG_PATH=/app/config/prod.yaml

CMD ["./url-shortener"]
