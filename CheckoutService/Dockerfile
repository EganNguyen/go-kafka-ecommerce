FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o checkout-service ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/checkout-service .

EXPOSE 8080

CMD ["./checkout-service"]
