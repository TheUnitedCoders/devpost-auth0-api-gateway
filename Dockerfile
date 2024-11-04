FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o gateway cmd/gateway/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install ca-certificates -y && update-ca-certificates

COPY --from=builder /app/gateway /gateway

EXPOSE 7070

ENTRYPOINT ["./gateway"]
