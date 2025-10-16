PG_ADDRESS=postgres://postgres@127.0.0.1/gochat
CASSANDRA_ADDRESS=cassandra://127.0.0.1/gochat?x-multi-statement=true

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
	go install github.com/swaggo/swag/v2/cmd/swag@latest

run:
	go run ./cmd/api

run_ws:
	go run ./cmd/ws

citus_up:
	docker compose -p gochat up --scale citus-worker=3 -d

migrate: migrate_pg migrate_scylla

migrate_down: migrate_pg_down migrate_scylla_down

migrate_scylla:
	migrate -database $(CASSANDRA_ADDRESS) -path ./db/cassandra up

migrate_pg:
	migrate -database $(PG_ADDRESS) -path ./db/postgres up

migrate_scylla_down:
	migrate -database $(CASSANDRA_ADDRESS) -path ./db/cassandra down

migrate_scylla_rollback:
	migrate -database $(CASSANDRA_ADDRESS) -path ./db/cassandra down 1

migrate_pg_down:
	migrate -database $(PG_ADDRESS) -path ./db/postgres down

migrate_pg_rollback:
	migrate -database $(PG_ADDRESS) -path ./db/postgres down 1

add_migration_postgres:
	migrate create -ext sql -dir db/postgres -seq $(name)

add_migration_cassandra:
	migrate create -ext cql -dir db/cassandra -seq $(name)

swag:
	swag fmt
	swag init -g doc.go \
	  	--v3.1 \
		--o ./docs/api \
		--ot json \
		--parseDependency \
		--parseInternal \
		--collectionFormat multi

client: js_client go_client

js_client:
	docker run --rm -v "./:/local/" mirror.gcr.io/openapitools/openapi-generator-cli:v7.12.0 \
			generate -i /local/docs/api/swagger.json -g typescript-axios -o /local/clients/api/jsclient --additional-properties=useSingleRequestParameter=true,withInterfaces=false,supportsES6=true

go_client:
	docker run --rm -v "./:/local/" mirror.gcr.io/openapitools/openapi-generator-cli:v7.12.0 \
			generate -i /local/docs/api/swagger.json -g go -o /local/clients/api/goclient --additional-properties=useSingleRequestParameter=true --package-name goclient --git-user-id FlameInTheDark --git-repo-id gochat/clients/api/goclient

setup: tools up migrate

.PHONY: setup

# Dev tools
rebuild_all: rebuild_api rebuild_auth rebuild_indexer rebuild_ws

rebuild_api:
	docker compose down api
	docker compose up -d --no-deps --build api

rebuild_auth:
	docker compose down auth
	docker compose up -d --no-deps --build auth

rebuild_ws:
	docker compose down ws
	docker compose up -d --no-deps --build ws

rebuild_indexer:
	docker compose down indexer
	docker compose up -d --no-deps --build indexer

rebuild_attachments:
	docker compose down attachments
	docker compose up -d --no-deps --build attachments