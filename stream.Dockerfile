FROM golang:1.26.2 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o stream ./cmd/stream

FROM alpine:latest
WORKDIR /dist
COPY --from=builder /build/stream .
CMD ["./stream"]
