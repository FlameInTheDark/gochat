FROM golang:1.24.4 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ws ./cmd/ws

FROM alpine:latest
WORKDIR /dist
COPY --from=builder /build/ws .
CMD ["./ws"]
