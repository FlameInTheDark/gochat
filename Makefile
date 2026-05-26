-include .env

COMPOSE_PROJECT_NAME?=gochat
YUGABYTE_USER?=yugabyte
YUGABYTE_PASSWORD?=yugabyte
YUGABYTE_DB?=gochat
YUGABYTE_HOST?=127.0.0.1
YUGABYTE_PORT?=5433
YUGABYTE_COMPOSE_HOST?=yugabyte
YUGABYTE_ADDRESS?=postgres://$(YUGABYTE_USER):$(YUGABYTE_PASSWORD)@$(YUGABYTE_HOST):$(YUGABYTE_PORT)/$(YUGABYTE_DB)?sslmode=disable
YUGABYTE_COMPOSE_ADDRESS?=postgres://$(YUGABYTE_USER):$(YUGABYTE_PASSWORD)@$(YUGABYTE_COMPOSE_HOST):5433/$(YUGABYTE_DB)?sslmode=disable
PG_ADDRESS?=$(YUGABYTE_ADDRESS)
CASSANDRA_KEYSPACE?=gochat
CASSANDRA_HOST?=127.0.0.1
CASSANDRA_COMPOSE_HOST?=scylla
CASSANDRA_ADDRESS?=cassandra://$(CASSANDRA_HOST)/$(CASSANDRA_KEYSPACE)?x-multi-statement=true
CASSANDRA_COMPOSE_ADDRESS?=cassandra://$(CASSANDRA_COMPOSE_HOST)/$(CASSANDRA_KEYSPACE)?x-multi-statement=true
MIGRATION_IMAGE?=gochat-migrations:local
MIGRATION_NETWORK?=$(COMPOSE_PROJECT_NAME)_default
MIGRATE_VERSION?=v4.19.1
MIGRATE_DOCKER_IMAGE?=migrate/migrate:$(MIGRATE_VERSION)

up:
	docker compose up -d
	docker compose exec scylla bash ./init-scylladb.sh

scylla_init:
	docker compose exec scylla bash ./init-scylladb.sh

down:
	docker compose down

tools:
	go install -tags "postgres cassandra" github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)
	go install github.com/swaggo/swag/v2/cmd/swag@latest

lint:
	golangci-lint run --timeout 5m

build_migration_image:
	docker build --build-arg MIGRATE_VERSION=$(MIGRATE_VERSION) -f migration.Dockerfile -t $(MIGRATION_IMAGE) .

migrate_image: build_migration_image
	docker run --rm \
		--network $(MIGRATION_NETWORK) \
		-e YUGABYTE_ADDRESS="$(YUGABYTE_COMPOSE_ADDRESS)" \
		-e CASSANDRA_ADDRESS="$(CASSANDRA_COMPOSE_ADDRESS)" \
		$(MIGRATION_IMAGE)

migrate_image_down: build_migration_image
	docker run --rm \
		--network $(MIGRATION_NETWORK) \
		-e YUGABYTE_ADDRESS="$(YUGABYTE_COMPOSE_ADDRESS)" \
		-e CASSANDRA_ADDRESS="$(CASSANDRA_COMPOSE_ADDRESS)" \
		$(MIGRATION_IMAGE) down

migrate_image_yugabyte: build_migration_image
	docker run --rm \
		--network $(MIGRATION_NETWORK) \
		-e MIGRATION_SCOPE=yugabyte \
		-e YUGABYTE_ADDRESS="$(YUGABYTE_COMPOSE_ADDRESS)" \
		$(MIGRATION_IMAGE)

migrate_image_scylla: build_migration_image
	docker run --rm \
		--network $(MIGRATION_NETWORK) \
		-e MIGRATION_SCOPE=cassandra \
		-e CASSANDRA_ADDRESS="$(CASSANDRA_COMPOSE_ADDRESS)" \
		$(MIGRATION_IMAGE)

migrate_image_scylla_down: build_migration_image
	docker run --rm \
		--network $(MIGRATION_NETWORK) \
		-e MIGRATION_SCOPE=cassandra \
		-e CASSANDRA_ADDRESS="$(CASSANDRA_COMPOSE_ADDRESS)" \
		$(MIGRATION_IMAGE) down

migrate_image_scylla_rollback: build_migration_image
	docker run --rm \
		--network $(MIGRATION_NETWORK) \
		-e MIGRATION_SCOPE=cassandra \
		-e CASSANDRA_ADDRESS="$(CASSANDRA_COMPOSE_ADDRESS)" \
		$(MIGRATION_IMAGE) down 1

run:
	go run ./cmd/api

run_ws:
	go run ./cmd/ws

run_embedder:
	go run ./cmd/embedder

yugabyte_up:
	docker compose up -d yugabyte yugabyte-init

migrate: migrate_yugabyte migrate_scylla

migrate_down: migrate_yugabyte_down migrate_scylla_down

migrate_yugabyte:
	docker run --rm \
		--network $(MIGRATION_NETWORK) \
		-v "./migration:/migrations:ro" \
		$(MIGRATE_DOCKER_IMAGE) \
		-database "$(YUGABYTE_COMPOSE_ADDRESS)" \
		-path /migrations/yugabyte up

migrate_yugabyte_down:
	docker run --rm \
		--network $(MIGRATION_NETWORK) \
		-v "./migration:/migrations:ro" \
		$(MIGRATE_DOCKER_IMAGE) \
		-database "$(YUGABYTE_COMPOSE_ADDRESS)" \
		-path /migrations/yugabyte down

migrate_yugabyte_rollback:
	docker run --rm \
		--network $(MIGRATION_NETWORK) \
		-v "./migration:/migrations:ro" \
		$(MIGRATE_DOCKER_IMAGE) \
		-database "$(YUGABYTE_COMPOSE_ADDRESS)" \
		-path /migrations/yugabyte down 1

migrate_scylla:
	docker run --rm \
		--network $(MIGRATION_NETWORK) \
		-v "./migration:/migrations:ro" \
		$(MIGRATE_DOCKER_IMAGE) \
		-database "$(CASSANDRA_COMPOSE_ADDRESS)" \
		-path /migrations/cassandra up

migrate_scylla_down:
	docker run --rm \
		--network $(MIGRATION_NETWORK) \
		-v "./migration:/migrations:ro" \
		$(MIGRATE_DOCKER_IMAGE) \
		-database "$(CASSANDRA_COMPOSE_ADDRESS)" \
		-path /migrations/cassandra down

migrate_scylla_rollback:
	docker run --rm \
		--network $(MIGRATION_NETWORK) \
		-v "./migration:/migrations:ro" \
		$(MIGRATE_DOCKER_IMAGE) \
		-database "$(CASSANDRA_COMPOSE_ADDRESS)" \
		-path /migrations/cassandra down 1

migrate_yugabyte_local:
	migrate -database $(YUGABYTE_ADDRESS) -path ./migration/yugabyte up

migrate_scylla_local:
	migrate -database $(CASSANDRA_ADDRESS) -path ./migration/cassandra up

add_migration_yugabyte:
	migrate create -ext sql -dir migration/yugabyte -seq $(name)

add_migration_cassandra:
	migrate create -ext cql -dir migration/cassandra -seq $(name)

swag:
	swag fmt
	swag init -g doc.go \
		--o ./docs/api \
		--ot json \
		--parseDependency \
		--parseInternal \
		--collectionFormat multi

client: js_client go_client

js_client:
	docker run --rm -v "./:/local/" mirror.gcr.io/openapitools/openapi-generator-cli:v7.12.0 \
			generate -i /local/docs/api/swagger.json -g typescript-axios -o /local/clients/api/jsclient \
			--additional-properties=useSingleRequestParameter=true,withInterfaces=false,supportsES6=true \
			--type-mappings=integer+int64=string \
			--skip-validate-spec

go_client:
	docker run --rm -v "./:/local/" mirror.gcr.io/openapitools/openapi-generator-cli:v7.12.0 \
			generate -i /local/docs/api/swagger.json -g go -o /local/clients/api/goclient --additional-properties=useSingleRequestParameter=true --package-name goclient --git-user-id FlameInTheDark --git-repo-id  gochat/clients/api/goclient --skip-validate-spec

setup: tools up migrate

.PHONY: setup tools lint build_migration_image migrate_image migrate_image_down migrate_image_yugabyte migrate_image_scylla migrate_image_scylla_down migrate_image_scylla_rollback run run_ws run_embedder yugabyte_up migrate migrate_down migrate_yugabyte migrate_yugabyte_down migrate_yugabyte_rollback migrate_scylla migrate_scylla_down migrate_scylla_rollback migrate_yugabyte_local migrate_scylla_local add_migration_yugabyte add_migration_cassandra rebuild_all rebuild_api rebuild_auth rebuild_ws rebuild_bot_services rebuild_botapi rebuild_botws rebuild_botrouter rebuild_indexer rebuild_attachments rebuild_sfu rebuild_webhook rebuild_embedder rebuild_telemetry_gateway

# Dev tools
rebuild_all: rebuild_api rebuild_auth rebuild_indexer rebuild_embedder rebuild_ws rebuild_bot_services

rebuild_api:
	docker compose down api
	docker compose up -d --no-deps --build api

rebuild_auth:
	docker compose down auth
	docker compose up -d --no-deps --build auth

rebuild_ws:
	docker compose down ws
	docker compose up -d --no-deps --build ws

rebuild_bot_services: rebuild_botapi rebuild_botws rebuild_botrouter

rebuild_botapi:
	docker compose down botapi
	docker compose up -d --no-deps --build botapi

rebuild_botws:
	docker compose down botws
	docker compose up -d --no-deps --build botws

rebuild_botrouter:
	docker compose down botrouter
	docker compose up -d --no-deps --build botrouter

rebuild_indexer:
	docker compose down indexer
	docker compose up -d --no-deps --build indexer

rebuild_attachments:
	docker compose down attachments
	docker compose up -d --no-deps --build attachments

rebuild_sfu:
	docker compose down sfu
	docker compose up -d --no-deps --build sfu

rebuild_webhook:
	docker compose down webhook
	docker compose up -d --no-deps --build webhook

rebuild_embedder:
	docker compose down embedder
	docker compose up -d --no-deps --build embedder

rebuild_telemetry_gateway:
	docker compose down telemetry-gateway
	docker compose up -d --no-deps --build telemetry-gateway
