FROM golang:1.25 AS builder

WORKDIR /app
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -p=1 -o room-service ./cmd/roomService

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/room-service .

CMD ["./room-service"]