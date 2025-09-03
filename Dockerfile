FROM golang:alpine AS builder
WORKDIR /app

COPY go.mod .
# COPY go.sum .

RUN go mod tidy

COPY . .

RUN go build -o main main.go
# ENV HOST_IP=$(hostname -I | cut -d' ' -f1)

FROM alpine

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]