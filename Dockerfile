# syntax=docker/dockerfile:1

FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/node ./cmd/node/main.go
RUN go build -o /bin/gateway ./cmd/gateway/main.go

FROM alpine:latest

COPY --from=builder /bin/node /bin/node
COPY --from=builder /bin/gateway /bin/gateway

EXPOSE 50051 50052 50053 8080 9101 9102 9103

CMD ["/bin/node"]