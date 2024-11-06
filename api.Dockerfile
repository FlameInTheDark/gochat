FROM golang:1.23.1 as builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api ./cmd/api

FROM alpine:latest
WORKDIR /dist
COPY --from=builder /build/api .
COPY --from=builder /build/docs/api docs/api
CMD ["./api"]
