FROM golang:1.22-alpine AS builder

WORKDIR /app

# Initialize the module inside Docker if not using go.mod locally yet,
# or better yet, since we have go.mod in other folders, let's create a minimal one.
COPY go.mod ./
RUN go mod tidy

COPY . .

RUN go build -o main ./main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
