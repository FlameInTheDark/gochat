FROM golang:1.26.2 AS builder

ARG MIGRATE_VERSION=v4.19.1

# YugabyteDB YSQL uses the PostgreSQL wire protocol, so golang-migrate's
# postgres driver is still required even though only YugabyteDB SQL migrations are shipped.
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
