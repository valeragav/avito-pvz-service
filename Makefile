# Загружаем .env, если он существует
ifneq (,$(wildcard .env))
	include .env
	export
endif

PROJECT_NAME=avito-pvz-service

## help: Show this help message with available commands
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## fast-start: Run the application quickly without building a binary
.PHONY: fast-start
fast-start:
	go run cmd/app/main.go

## start: Build and run the application binary
.PHONY: start
start:
	go build -o build/package/$(PROJECT_NAME) cmd/app/main.go
	build/package/$(PROJECT_NAME)

## test: Run all unit tests with verbose output
.PHONY: test
test:
	go test -v -fullpath=true -timeout 30s ./...

## lint: Run code linter to check for errors and style issues, make sure to install golangci-lint
.PHONY: lint
lint:
	golangci-lint run

## swagger-init: Generate Swagger API documentation,  make sure to install swag CLI
.PHONY: swagger-init
swagger-init:
	swag init -g internal/api/swagger.go -o docs

## gen: Run go generate on all packages (e.g., code generation)
.PHONY: gen
gen:
	go generate ./...

# Migration
DOCKER_NETWORK=avito-pvz-service_avito-pvz-service_network
DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@postgres:$(DB_INTERNAL_PORT)/$(DB_NAME)?$(DB_OPTION)

MIGRATE_RUN=docker run --rm \
	$(if $(DOCKER_NETWORK),--network $(DOCKER_NETWORK),) \
	-v $(PWD)/migrations:/migrations \
	migrate/migrate:v4.19.1 \
	-path=/migrations \
	-database "$(DATABASE_URL)"

## create-migration: Create an empty migration
.PHONY: create-migration
create-migration:
	@read -p "Enter migration name: " NAME; \
	docker run --rm \
		-v $(PWD)/migrations:/migrations \
		migrate/migrate:v4.19.1 \
		create -ext sql -dir /migrations -seq $$NAME

## migrate-up: Migration up
.PHONY: migrate-up
migrate-up:
	$(MIGRATE_RUN) up

## migrate-down: Migration down
.PHONY: migrate-down
migrate-down:
	@read -p "Number of migrations to rollback (default: 1): " NUM; \
	NUM=$${NUM:-1}; \
	$(MIGRATE_RUN) down $$NUM

## migrate-version: Migration version
.PHONY: migrate-version
migrate-version:
	$(MIGRATE_RUN) version



PROTO=api/v1/proto/*.proto
OUT=internal/api/grpc/gen

proto:
	protoc -I api/v1/proto \
		$(PROTO) \
		--go_out=$(OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT) --go-grpc_opt=paths=source_relative



PROTO_DIR = api/v1/proto
OUT_DIR   = internal/api/grpc/gen/v1

proto-1:
	protoc -I $(PROTO_DIR) $(PROTO_DIR)/*.proto \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		--plugin=protoc-gen-validate=/home/halon/go/bin/protoc-gen-validate \
		--validate_out="lang=go,paths=source_relative:$(OUT_DIR)"