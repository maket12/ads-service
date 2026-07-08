FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ ./pkg/
COPY userservice/ ./userservice/

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/userservice ./userservice/cmd/app/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/userservice /app/userservice

EXPOSE 50052
CMD ["/app/userservice"]
