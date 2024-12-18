include .envrc
MIGRATION_PATH = ./cmd/migrate/migrations

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(name)" ]; then echo "Uso: make migrate-create name=<migration_name>"; exit 1; fi
	@migrate create -seq -ext sql -dir $(MIGRATION_PATH) $(name)

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) up

.PHONY: migrate-down
migrate-down:
	@if [ -z "$(steps)" ]; then steps=all; fi
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) down $(steps)

.PHONY: migrate-force
migrate-force:
	@if [ -z "$(version)" ]; then echo "Uso: make migrate-force version=<version>"; exit 1; fi
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) force $(version)

.PHONY: seed
seed:
	@go run cmd/migrate/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt

