FROM golang:1.25.1 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o attachments ./cmd/attachments

FROM alpine:latest
WORKDIR /dist
RUN apk add --no-cache ca-certificates ffmpeg
COPY --from=builder /build/attachments .
CMD ["./attachments"]

