FROM golang:1.23-alpine AS builder

RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd/
COPY pkg/ ./pkg/

RUN go build -o /app/ssh-proxy cmd/main.go

FROM alpine:latest
COPY --from=builder /app/ssh-proxy /app/ssh-proxy