FROM golang:1.25.8 AS builder

ARG MIGRATE_VERSION=v4.19.1

RUN CGO_ENABLED=0 go install -tags "postgres cassandra" github.com/golang-migrate/migrate/v4/cmd/migrate@${MIGRATE_VERSION}

FROM alpine:3.22

WORKDIR /migrations

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY migration /migrations
COPY init/run-migrations.sh /usr/local/bin/run-migrations
RUN sed -i 's/\r$//' /usr/local/bin/run-migrations && chmod +x /usr/local/bin/run-migrations

ENTRYPOINT ["/usr/local/bin/run-migrations"]
CMD ["up"]
