FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o tinycdn ./cmd

FROM alpine:3.20.0

WORKDIR /app

COPY --from=builder /app/tinycdn .

EXPOSE 8080

CMD ["./tinycdn"]
