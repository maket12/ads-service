FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg/
COPY internal/ ./internal/
COPY "cmd/" "./cmd/"

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/authservice ./cmd/app/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/authservice /app/authservice

EXPOSE 50051

CMD ["/app/authservice"]
