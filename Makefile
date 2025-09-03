up:
	docker compose up -d
	docker compose exec scylla bash ./init-scylladb.sh
	docker compose -p gochat up --scale citus-worker=3 -d

scylla_init:
	docker compose exec scylla bash ./init-scylladb.sh

down:
	docker compose down

tools:
	go install github.com/db-journey/journey/v2
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
	journey --url cassandra://127.0.0.1/gochat --path ./db/migrations migrate up

migrate_pg:
	journey --url postgres://postgres@127.0.0.1/gochat --path ./db/pgmigrations migrate up

migrate_scylla_down:
	journey --url cassandra://127.0.0.1/gochat --path ./db/migrations migrate down

migrate_pg_down:
	journey --url postgres://postgres@127.0.0.1/gochat --path ./db/pgmigrations migrate down

swag:
	swag fmt
	swag init --ot json -o ./docs/api -g cmd/api/main.go --parseDependency

# PostgreSQL database management
create_pg_db:
	docker compose exec citus-master createdb -U postgres --if-not-exists gochat

setup: tools up create_pg_db migrate

.PHONY: setup