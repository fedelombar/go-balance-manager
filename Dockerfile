FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o go-balance-manager

FROM alpine:3.17

WORKDIR /app

RUN apk add --no-cache postgresql-client

COPY --from=builder /app/go-balance-manager /app/go-balance-manager

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
