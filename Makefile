# Загружаем .env, если он существует
ifneq (,$(wildcard .env))
	include .env
	export
endif

PROJECT_NAME=avito-pvz-service

DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@postgres:$(DB_INTERNAL_PORT)/$(DB_NAME)?$(DB_OPTION)

DOCKER_NETWORK=avito-pvz-service_avito-pvz-service_network

MIGRATE_RUN=docker run --rm \
	$(if $(DOCKER_NETWORK),--network $(DOCKER_NETWORK),) \
	-v $(PWD)/migrations:/migrations \
	migrate/migrate:v4.19.1 \
	-path=/migrations \
	-database "$(DATABASE_URL)"

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## init: used to initialize the Go project, tidy, docker, migration, build and deploy
.PHONY: init
init:
	docker compose up -d
	go build -o build/package/$(PROJECT_NAME) cmd/api/main.go

## init-docker: used to initialize the Go project docker, tidy, docker, migration, build and deploy
.PHONY: init-docker
init-docker:
	go build -o build/package/$(PROJECT_NAME) cmd/api/main.go
	build/package/$(PROJECT_NAME)

## deploy: executing the deployment command
.PHONY: swagger-init
swagger-init:
	swag init -g cmd/api/main.go -o docs

## fast-start: quick launch
.PHONY: fast-start
fast-start:
	go run cmd/api/main.go

## start: build start
.PHONY: start
start:
	go build -o build/package/$(PROJECT_NAME) cmd/api/main.go
	build/package/$(PROJECT_NAME)

## gen: generate code
.PHONY: gen
gen:
	go generate ./...

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