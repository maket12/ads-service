FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY api/proto/generated ./api/proto/generated
COPY internal/ ./internal/
COPY "cmd/" "./cmd/"

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/userservice ./cmd/app/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/userservice /app/userservice

EXPOSE 50052

CMD ["/app/userservice"]
