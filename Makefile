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

## gen: Run go generate on all packages (e.g., code generation)
.PHONY: gen
gen:
	go generate ./...

# Migration Docker
DOCKER_NETWORK=avito-pvz-service_avito-pvz-service_network
DATABASE_URL="postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_INTERNAL_PORT)/$(DB_NAME)?$(DB_OPTION)"

MIGRATE_RUN=docker run --rm \
	$(if $(DOCKER_NETWORK),--network $(DOCKER_NETWORK),) \
	-v $(PWD)/migrations:/migrations \
	migrate/migrate:v4.19.1 \
	-path=/migrations \
	-database $(DATABASE_URL)

## migration-docker-create: Create an empty migration
.PHONY: migration-docker-create
migration-docker-create:
	@read -p "Enter migration name: " NAME; \
	docker run --rm \
		-v $(PWD)/migrations:/migrations \
		migrate/migrate:v4.19.1 \
		create -ext sql -dir /migrations -seq $$NAME

## migrate-docker-up: Migration up
.PHONY: migrate-docker-up
migrate-docker-up:
	$(MIGRATE_RUN) up

## migrate-docker-down: Migration down
.PHONY: migrate-docker-down
migrate-docker-down:
	@read -p "Number of migrations to rollback (default: 1): " NUM; \
	NUM=$${NUM:-1}; \
	$(MIGRATE_RUN) down $$NUM

## migrate-docker-version: Migration version
.PHONY: migrate-docker-version
migrate-docker-version:
	$(MIGRATE_RUN) version

LOCAL_BIN:=$(CURDIR)/bin

## bin-deps: Install all necessary binary dependencies
.PHONY: bin-deps
bin-deps:
	$(info Installing binary dependencies...)
	@mkdir -p $(LOCAL_BIN)

	@tmp_dir=$$(mktemp -d); \
	git clone --depth 1 --branch v2.10.1 https://github.com/golangci/golangci-lint.git $$tmp_dir; \
	cd $$tmp_dir/cmd/golangci-lint; \
	go build -o $(LOCAL_BIN)/golangci-lint; \
	cd -; \
	rm -rf $$tmp_dir

	GOBIN=$(LOCAL_BIN) go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	GOBIN=$(LOCAL_BIN) go install github.com/swaggo/swag/cmd/swag@latest
	GOBIN=$(LOCAL_BIN) go install go.uber.org/mock/mockgen@latest
	GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@latest
	GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

## lint: Run code linter
.PHONY: lint
lint:
	$(LOCAL_BIN)/golangci-lint run 

## swagger-init: Generate Swagger API documentation
.PHONY: swagger-init
swagger-init:
	$(LOCAL_BIN)/swag init -g internal/api/swagger.go -o docs

PROTO_DIR = ./api/v1/proto
OUT_DIR   = ./internal/api/grpc/gen/v1

## all-generate-proto: Install dependencies and generate proto code
.PHONY: all-generate-proto
all-generate-proto: bin-deps generate-proto

## generate-proto: Generate gRPC and Protobuf code with validation
.PHONY: generate-proto
generate-proto:
	mkdir -p $(OUT_DIR)
	protoc -I $(PROTO_DIR) $(PROTO_DIR)/*.proto \
		--plugin=protoc-gen-go=$(LOCAL_BIN)/protoc-gen-go --go_out=$(OUT_DIR) --go_opt=paths=source_relative\
		--plugin=protoc-gen-go-grpc=$(LOCAL_BIN)/protoc-gen-go-grpc --go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		--plugin=protoc-gen-validate=$(LOCAL_BIN)/protoc-gen-validate --validate_out="lang=go,paths=source_relative:$(OUT_DIR)"

## migrate-create: Create a new migration with local migrate binary
.PHONY: migrate-create
migrate-create:
	@test -n "$(NAME)" || (echo "Error: NAME variable is required"; exit 1)
	$(LOCAL_BIN)/migrate create -ext sql -dir ${CURDIR}/migrations -seq $(NAME)

## migrate-up: Apply migrations using local migrate binary
.PHONY: migrate-up
migrate-up:
	$(LOCAL_BIN)/migrate -path=${CURDIR}/migrations -database=$(DATABASE_URL) up

## migrate-down: Rollback migrations using local migrate binary
.PHONY: migrate-down
migrate-down:
	@read -p "Number of migrations to rollback (default: 1): " NUM; NUM=$${NUM:-1}; \
	$(LOCAL_BIN)/migrate -path=${CURDIR}/migrations -database=$(DATABASE_URL) down $(NUM)

## migrate-version: Show current migration version
.PHONY: migrate-version
migrate-version:
	$(LOCAL_BIN)/migrate -path=${CURDIR}/migrations -database=$(DATABASE_URL) version