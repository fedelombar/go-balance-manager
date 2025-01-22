FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bin/go-balance-manager

FROM alpine:3.17

WORKDIR /app

COPY --from=builder /app/bin/go-balance-manager /app/go-balance-manager

EXPOSE 8080

ENV DB_HOST=db
ENV DB_PORT=5432
ENV DB_USER=golang_user
ENV DB_PASS=golang_pass
ENV DB_NAME=golang_db
ENV APP_ADDR=:8080

CMD ["./go-balance-manager", "-addr=:8080"]
