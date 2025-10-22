FROM golang:1.25.1 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o webhook ./cmd/webhook

FROM alpine:latest
WORKDIR /dist
COPY --from=builder /build/webhook .
CMD ["./webhook"]
