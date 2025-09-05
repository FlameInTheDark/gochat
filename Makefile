up:
	docker compose up -d
	docker compose exec scylla bash ./init-scylladb.sh
	docker compose -p gochat up --scale citus-worker=3 -d

scylla_init:
	docker compose exec scylla bash ./init-scylladb.sh

down:
	docker compose down

tools:
	go install -tags "postgres cassandra" github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/swaggo/swag/cmd/swag@latest

run:
	go run ./cmd/api

run_ws:
	go run ./cmd/ws

citus_up:
	docker compose -p gochat up --scale citus-worker=3 -d

migrate: migrate_pg migrate_scylla

migrate_down: migrate_pg_down migrate_scylla_down

migrate_scylla:
	migrate -database cassandra://127.0.0.1/gochat?x-multi-statement=true -path ./db/cassandra up

migrate_pg:
	migrate -database postgres://postgres@127.0.0.1/gochat -path ./db/postgres up

migrate_scylla_down:
	migrate -database cassandra://127.0.0.1/gochat?x-multi-statement=true -path ./db/cassandra down

migrate_pg_down:
	migrate -database postgres://postgres@127.0.0.1/gochat -path ./db/postgres down

swag:
	swag fmt
	swag init --ot json -o ./docs/api -g cmd/api/main.go --parseDependency

setup: tools up migrate

.PHONY: setup