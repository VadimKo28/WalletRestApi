FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o wallet-app ./cmd

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/wallet-app /app/wallet-app

ENV GIN_MODE=release

EXPOSE 3010

CMD ["/app/wallet-app"]


