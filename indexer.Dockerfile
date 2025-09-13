FROM golang:1.25.1 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o indexer ./cmd/indexer

FROM alpine:latest
WORKDIR /dist
COPY --from=builder /build/indexer .
CMD ["./indexer"]
