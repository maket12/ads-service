FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ ./pkg/
COPY adservice/ ./adservice/

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/adservice ./adservice/cmd/app/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/adservice /app/adservice

EXPOSE 50053
CMD ["/app/adservice"]
