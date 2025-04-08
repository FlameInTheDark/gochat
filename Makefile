up:
	docker compose up -d
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

migrate:
	journey --url cassandra://127.0.0.1/gochat --path ./db/migrations migrate up

migrate_down:
	journey --url cassandra://127.0.0.1/gochat --path ./db/migrations migrate down

swag:
	swag fmt
	swag init --ot json -o ./docs/api -g cmd/api/main.go --parseDependency

setup: tools up migrate

# Helm Chart Management
# Variables
HELM_CHART_PATH ?= ./gochat-chart
HELM_RELEASE_NAME ?= gochat
HELM_NAMESPACE ?= default
HELM_MIGRATIONS_SRC ?= ./db/migrations
HELM_MIGRATIONS_DEST ?= $(HELM_CHART_PATH)/db/migrations

# Windows-compatible path variables
WIN_SRC := $(subst /,\,${HELM_MIGRATIONS_SRC})
WIN_DEST := $(subst /,\,${HELM_MIGRATIONS_DEST})

.PHONY: helm-lint helm-template helm-install helm-upgrade helm-uninstall copy-migrations-to-chart clean-migrations-from-chart

copy-migrations-to-chart:
	@echo "Copying migrations to Helm chart directory..."
	@rem xcopy will create destination directories if /I is used and source has >1 file/dir
	@xcopy "${WIN_SRC}" "${WIN_DEST}\\" /E /I /Y /Q

clean-migrations-from-chart:
	@echo "Cleaning up copied migrations from Helm chart directory..."
	@rem Remove directory recursively and quietly, only if it exists
	@if exist "${WIN_DEST}" ( rmdir /S /Q "${WIN_DEST}" )

helm-lint:
	$(MAKE) copy-migrations-to-chart
	helm lint $(HELM_CHART_PATH)
	$(MAKE) clean-migrations-from-chart

helm-template:
	$(MAKE) copy-migrations-to-chart
	helm template $(HELM_RELEASE_NAME) $(HELM_CHART_PATH) --namespace $(HELM_NAMESPACE) > rendered-manifest.yaml
	$(MAKE) clean-migrations-from-chart
	@echo "Rendered manifest saved to rendered-manifest.yaml"

helm-install:
	$(MAKE) copy-migrations-to-chart
	helm install $(HELM_RELEASE_NAME) $(HELM_CHART_PATH) --namespace $(HELM_NAMESPACE) --create-namespace
	$(MAKE) clean-migrations-from-chart

helm-upgrade:
	$(MAKE) copy-migrations-to-chart
	helm upgrade --install $(HELM_RELEASE_NAME) $(HELM_CHART_PATH) --namespace $(HELM_NAMESPACE) --create-namespace
	$(MAKE) clean-migrations-from-chart

helm-uninstall:
	helm uninstall $(HELM_RELEASE_NAME) --namespace $(HELM_NAMESPACE)
