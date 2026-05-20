YUGABYTE_ADDRESS?=postgres://yugabyte:yugabyte@127.0.0.1:5433/gochat?sslmode=disable
CITUS_ADDRESS?=postgres://postgres@127.0.0.1:5432/gochat?sslmode=disable
PG_ADDRESS?=$(YUGABYTE_ADDRESS)
CASSANDRA_ADDRESS?=cassandra://127.0.0.1/gochat?x-multi-statement=true
MIGRATION_IMAGE?=gochat-migrations:local
MIGRATE_VERSION?=v4.19.1
VOYAGER_IMAGE?=software.yugabyte.com/yugabytedb/yb-voyager
VOYAGER_DOCKER_NETWORK?=gochat_default
VOYAGER?=docker run --rm --network $(VOYAGER_DOCKER_NETWORK) -v "./.data/voyager:/voyager" $(VOYAGER_IMAGE) yb-voyager
VOYAGER_EXPORT_DIR?=/voyager/gochat
VOYAGER_EXPORT_HOST_DIR?=.data/voyager/gochat
VOYAGER_SOURCE_SCHEMA?=public
VOYAGER_EXCLUDE_TABLES?=schema_migrations
CITUS_HOST?=citus-master
CITUS_PORT?=5432
CITUS_USER?=postgres
CITUS_PASSWORD?=postgres
CITUS_DB?=gochat
YUGABYTE_HOST?=yugabyte
YUGABYTE_PORT?=5433
YUGABYTE_USER?=yugabyte
YUGABYTE_PASSWORD?=yugabyte
YUGABYTE_DB?=gochat
YUGABYTE_VERIFY_IMAGE?=golang:1.26.2
YUGABYTE_VERIFY_SOURCE_DSN?=postgres://$(CITUS_USER):$(CITUS_PASSWORD)@$(CITUS_HOST):$(CITUS_PORT)/$(CITUS_DB)?sslmode=disable
YUGABYTE_VERIFY_TARGET_DSN?=postgres://$(YUGABYTE_USER):$(YUGABYTE_PASSWORD)@$(YUGABYTE_HOST):$(YUGABYTE_PORT)/$(YUGABYTE_DB)?sslmode=disable

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
		-e YUGABYTE_ADDRESS="$(YUGABYTE_ADDRESS)" \
		-e CASSANDRA_ADDRESS="$(CASSANDRA_ADDRESS)" \
		$(MIGRATION_IMAGE)

migrate_image_down: build_migration_image
	docker run --rm \
		-e YUGABYTE_ADDRESS="$(YUGABYTE_ADDRESS)" \
		-e CASSANDRA_ADDRESS="$(CASSANDRA_ADDRESS)" \
		$(MIGRATION_IMAGE) down

migrate_image_yugabyte: build_migration_image
	docker run --rm \
		-e MIGRATION_SCOPE=yugabyte \
		-e YUGABYTE_ADDRESS="$(YUGABYTE_ADDRESS)" \
		$(MIGRATION_IMAGE)

migrate_image_scylla: build_migration_image
	docker run --rm \
		-e MIGRATION_SCOPE=cassandra \
		-e CASSANDRA_ADDRESS="$(CASSANDRA_ADDRESS)" \
		$(MIGRATION_IMAGE)

migrate_image_pg: build_migration_image
	docker run --rm \
		-e MIGRATION_SCOPE=citus \
		-e CITUS_ADDRESS="$(CITUS_ADDRESS)" \
		$(MIGRATION_IMAGE)

migrate_image_scylla_down: build_migration_image
	docker run --rm \
		-e MIGRATION_SCOPE=cassandra \
		-e CASSANDRA_ADDRESS="$(CASSANDRA_ADDRESS)" \
		$(MIGRATION_IMAGE) down

migrate_image_scylla_rollback: build_migration_image
	docker run --rm \
		-e MIGRATION_SCOPE=cassandra \
		-e CASSANDRA_ADDRESS="$(CASSANDRA_ADDRESS)" \
		$(MIGRATION_IMAGE) down 1

migrate_image_pg_down: build_migration_image
	docker run --rm \
		-e MIGRATION_SCOPE=citus \
		-e CITUS_ADDRESS="$(CITUS_ADDRESS)" \
		$(MIGRATION_IMAGE) down

migrate_image_pg_rollback: build_migration_image
	docker run --rm \
		-e MIGRATION_SCOPE=citus \
		-e CITUS_ADDRESS="$(CITUS_ADDRESS)" \
		$(MIGRATION_IMAGE) down 1

run:
	go run ./cmd/api

run_ws:
	go run ./cmd/ws

run_embedder:
	go run ./cmd/embedder

citus_up:
	docker compose --profile citus up --scale citus-worker=3 -d citus-master citus-manager citus-worker citus-init

yugabyte_up:
	docker compose up -d yugabyte yugabyte-init

migrate: migrate_yugabyte migrate_scylla

migrate_down: migrate_yugabyte_down migrate_scylla_down

migrate_yugabyte:
	migrate -database $(YUGABYTE_ADDRESS) -path ./migration/yugabyte up

migrate_yugabyte_down:
	migrate -database $(YUGABYTE_ADDRESS) -path ./migration/yugabyte down

migrate_yugabyte_rollback:
	migrate -database $(YUGABYTE_ADDRESS) -path ./migration/yugabyte down 1

migrate_scylla:
	migrate -database $(CASSANDRA_ADDRESS) -path ./migration/cassandra up

migrate_pg:
	migrate -database $(CITUS_ADDRESS) -path ./migration/postgres up

migrate_scylla_down:
	migrate -database $(CASSANDRA_ADDRESS) -path ./migration/cassandra down

migrate_scylla_rollback:
	migrate -database $(CASSANDRA_ADDRESS) -path ./migration/cassandra down 1

migrate_pg_down:
	migrate -database $(CITUS_ADDRESS) -path ./migration/postgres down

migrate_pg_rollback:
	migrate -database $(CITUS_ADDRESS) -path ./migration/postgres down 1

voyager_assess:
	powershell -NoProfile -Command "New-Item -ItemType Directory -Force -Path '$(VOYAGER_EXPORT_HOST_DIR)' | Out-Null"
	$(VOYAGER) assess-migration \
		--source-db-type postgresql \
		--source-db-host $(CITUS_HOST) \
		--source-db-port $(CITUS_PORT) \
		--source-db-user $(CITUS_USER) \
		--source-db-password "$(CITUS_PASSWORD)" \
		--source-db-name $(CITUS_DB) \
		--source-db-schema $(VOYAGER_SOURCE_SCHEMA) \
		--source-ssl-mode disable \
		--send-diagnostics false \
		--start-clean true \
		--export-dir $(VOYAGER_EXPORT_DIR) \
		-y

voyager_export:
	powershell -NoProfile -Command "New-Item -ItemType Directory -Force -Path '$(VOYAGER_EXPORT_HOST_DIR)' | Out-Null"
	$(VOYAGER) export data \
		--source-db-type postgresql \
		--source-db-host $(CITUS_HOST) \
		--source-db-port $(CITUS_PORT) \
		--source-db-user $(CITUS_USER) \
		--source-db-password "$(CITUS_PASSWORD)" \
		--source-db-name $(CITUS_DB) \
		--source-db-schema $(VOYAGER_SOURCE_SCHEMA) \
		--exclude-table-list "$(VOYAGER_EXCLUDE_TABLES)" \
		--source-ssl-mode disable \
		--send-diagnostics false \
		--disable-pb true \
		--start-clean true \
		--export-dir $(VOYAGER_EXPORT_DIR) \
		-y

voyager_import:
	powershell -NoProfile -Command "New-Item -ItemType Directory -Force -Path '$(VOYAGER_EXPORT_HOST_DIR)' | Out-Null"
	$(VOYAGER) import data \
		--target-db-host $(YUGABYTE_HOST) \
		--target-db-port $(YUGABYTE_PORT) \
		--target-db-user $(YUGABYTE_USER) \
		--target-db-password "$(YUGABYTE_PASSWORD)" \
		--target-db-name $(YUGABYTE_DB) \
		--target-ssl-mode disable \
		--send-diagnostics false \
		--disable-pb true \
		--start-clean true \
		--export-dir $(VOYAGER_EXPORT_DIR) \
		-y

voyager_status:
	powershell -NoProfile -Command "New-Item -ItemType Directory -Force -Path '$(VOYAGER_EXPORT_HOST_DIR)' | Out-Null"
	-$(VOYAGER) export data status --export-dir $(VOYAGER_EXPORT_DIR)
	-$(VOYAGER) import data status --export-dir $(VOYAGER_EXPORT_DIR)

yugabyte_verify:
	docker run --rm --network $(VOYAGER_DOCKER_NETWORK) -v "./:/src" -w /src $(YUGABYTE_VERIFY_IMAGE) \
		go run ./cmd/tools yugabyte verify \
		--source-dsn "$(YUGABYTE_VERIFY_SOURCE_DSN)" \
		--target-dsn "$(YUGABYTE_VERIFY_TARGET_DSN)"

add_migration_postgres:
	migrate create -ext sql -dir migration/postgres -seq $(name)

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

.PHONY: setup tools lint build_migration_image migrate_image migrate_image_down migrate_image_yugabyte migrate_image_scylla migrate_image_pg migrate_image_scylla_down migrate_image_scylla_rollback migrate_image_pg_down migrate_image_pg_rollback run run_ws run_embedder citus_up yugabyte_up migrate migrate_down migrate_yugabyte migrate_yugabyte_down migrate_yugabyte_rollback migrate_scylla migrate_pg migrate_scylla_down migrate_scylla_rollback migrate_pg_down migrate_pg_rollback voyager_assess voyager_export voyager_import voyager_status yugabyte_verify rebuild_all rebuild_api rebuild_auth rebuild_ws rebuild_indexer rebuild_attachments rebuild_sfu rebuild_webhook rebuild_embedder rebuild_telemetry_gateway

# Dev tools
rebuild_all: rebuild_api rebuild_auth rebuild_indexer rebuild_embedder rebuild_ws

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
