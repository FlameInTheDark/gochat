up:
	docker compose up -d
	docker compose exec scylla "./init-scylladb.sh"

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
