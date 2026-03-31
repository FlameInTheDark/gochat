FROM golang:1.25.8 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o telemetry-gateway ./cmd/telemetrygateway

FROM alpine:latest
WORKDIR /dist
COPY --from=builder /build/telemetry-gateway .
CMD ["./telemetry-gateway"]
