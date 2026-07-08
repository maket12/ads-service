FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY gateway/ ./gateway/
COPY pkg/ ./pkg/

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/gateway ./gateway/cmd/app/server.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bin/gateway .

EXPOSE 8080
CMD ["./gateway"]
