FROM golang:1.25 AS builder

WORKDIR /app
ENV CGO_ENABLED=0

COPY smart-campus-generated /smart-campus-generated
COPY api-gateway/go.mod api-gateway/go.sum ./
RUN go mod download

COPY api-gateway .

RUN go build -p=1 -o api-gateway ./cmd/apiGateway

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api-gateway .

CMD ["./api-gateway"]