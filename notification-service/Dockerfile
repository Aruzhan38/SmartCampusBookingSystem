FROM golang:1.25 AS builder

WORKDIR /app
ENV CGO_ENABLED=0

COPY smart-campus-generated /smart-campus-generated
COPY notification-service/go.mod notification-service/go.sum ./
RUN go mod download

COPY notification-service .

RUN go build -p=1 -o notification-service ./cmd/notificationService

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/notification-service .

CMD ["./notification-service"]