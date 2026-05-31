FROM golang:1.26.2 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o botrouter ./cmd/botrouter

FROM alpine:3.22
WORKDIR /dist
RUN apk add --no-cache ca-certificates && \
    addgroup -S gochat && \
    adduser -S -G gochat gochat
COPY --from=builder /build/botrouter .
RUN chown -R gochat:gochat /dist
USER gochat:gochat
CMD ["./botrouter"]
