# Load .env for targets that need DB_* (run from project root).
# Requires: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME in .env
-include .env
export

.PHONY: migrate-up migrate-down metabase metabase-down superset superset-down

migrate-up:
	go run ./cmd/migrate

migrate-down:
	@echo "Run: migrate -path database/migrations -database \"postgres://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=disable\" down"
	@echo "Install migrate CLI: brew install migrate"

metabase:
	docker compose up -d

metabase-down:
	docker compose down

superset:
	docker compose up -d superset

superset-down:
	docker compose stop superset

run:
	go run . 2>&1 | jq