FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -tags musl \
    -ldflags="-w -s -extldflags '-static'" \
    -o /app/app-bin \
    ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app
COPY --from=builder /app/app-bin /app/app-bin
COPY --from=builder /app/config  /app/config

ENV CONFIG_PATH=/app/config/config

EXPOSE 5000
USER app

ENTRYPOINT ["/app/app-bin"]
