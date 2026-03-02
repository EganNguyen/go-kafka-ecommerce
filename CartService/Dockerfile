FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o cart-service ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/cart-service .

EXPOSE 8080

CMD ["./cart-service"]
