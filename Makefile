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

.PHONY: helm-lint helm-template helm-install helm-upgrade helm-uninstall

helm-lint:
	helm lint $(HELM_CHART_PATH)

helm-template:
	helm template $(HELM_RELEASE_NAME) $(HELM_CHART_PATH) --namespace $(HELM_NAMESPACE) > rendered-manifest.yaml
	@echo "Rendered manifest saved to rendered-manifest.yaml"

helm-install:
	helm install $(HELM_RELEASE_NAME) $(HELM_CHART_PATH) --namespace $(HELM_NAMESPACE) --create-namespace

helm-upgrade:
	helm upgrade --install $(HELM_RELEASE_NAME) $(HELM_CHART_PATH) --namespace $(HELM_NAMESPACE) --create-namespace

helm-uninstall:
	helm uninstall $(HELM_RELEASE_NAME) --namespace $(HELM_NAMESPACE)
