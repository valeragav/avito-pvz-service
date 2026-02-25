# Загружаем .env, если он существует
ifneq (,$(wildcard .env))
	include .env
	export
endif

.DEFAULT_GOAL := help
LOCAL_BIN := $(CURDIR)/bin
PROJECT_NAME = avito-pvz-service

# Migration config
DOCKER_NETWORK = avito-pvz-service_avito-pvz-service_network
DATABASE_URL = "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_INTERNAL_PORT)/$(DB_NAME)?$(DB_OPTION)"

# Proto config
PROTO_DIR = ./api/v1/proto
OUT_DIR = ./internal/api/grpc/gen/v1

# Tool versions
GOLANGCI_LINT_VERSION = v2.10.1 # old version is being installed from go.mod 
SWAG_VERSION          := $(shell go list -m -f '{{.Version}}' github.com/swaggo/swag 2>/dev/null)
MIGRATE_VERSION       := $(shell go list -m -f '{{.Version}}' github.com/golang-migrate/migrate/v4 2>/dev/null)
MOCKGEN_VERSION       := $(shell go list -m -f '{{.Version}}' go.uber.org/mock 2>/dev/null)
PROTOC_GEN_GO_VERSION := $(shell go list -m -f '{{.Version}}' google.golang.org/protobuf 2>/dev/null)
PROTOC_GEN_GO_GRPC_VERSION    := $(shell go list -m -f '{{.Version}}' google.golang.org/grpc/cmd/protoc-gen-go-grpc 2>/dev/null)
PROTOC_GEN_VALIDATE_VERSION   := $(shell go list -m -f '{{.Version}}' github.com/envoyproxy/protoc-gen-validate 2>/dev/null)

## help: Show this help message with available commands
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## fast-start: Run the application quickly without building a binary
.PHONY: fast-start
fast-start:
	go run cmd/app/main.go

## seeder: Run utilities that fill the database with initial data.
.PHONY: seeder
seeder:
	docker compose --profile seeder up --build seeder --abort-on-container-exit

## start: Build and run the application binary
.PHONY: start
start: seeder
	./script/generate_secrets.sh
	docker compose up -d --wait

## test: Run all unit tests with verbose output
.PHONY: test
test:
	go test -race -fullpath=true \
		-coverprofile=coverage.out \
		-covermode=atomic \
		-coverpkg=./... \
		./...
	grep -vE '\.pb\.|_mock\.go|_gen\.go|/schema/|/logger/' coverage.out > coverage.filtered.out
	mv coverage.filtered.out coverage.out

## load-test: Run k6 load tests against the running application (requires app to be up)
.PHONY: load-test
load-test:
	docker compose --profile load-test up k6-load-test --abort-on-container-exit

## coverage: Show coverage report
.PHONY: coverage
coverage: test
	go tool cover -func=coverage.out

## coverage-ci: Show coverage report from existing coverage.out (used in CI)
.PHONY: coverage-ci
coverage-ci:
	go tool cover -func=coverage.out

## gen: Run go generate on all packages
.PHONY: gen
gen: $(LOCAL_BIN)/mockgen
	LOCAL_BIN=$(LOCAL_BIN) go generate ./...

## lint: Run code linter
.PHONY: lint
lint: $(LOCAL_BIN)/golangci-lint
	$(LOCAL_BIN)/golangci-lint run

## swagger-init: Generate Swagger API documentation
.PHONY: swagger-init
swagger-init: $(LOCAL_BIN)/swag
	$(LOCAL_BIN)/swag init -g internal/api/swagger.go -o api/v1/swagger

## generate-proto: Generate gRPC and Protobuf code with validation
.PHONY: generate-proto
generate-proto: $(LOCAL_BIN)/protoc-gen-go $(LOCAL_BIN)/protoc-gen-go-grpc $(LOCAL_BIN)/protoc-gen-validate
	mkdir -p $(OUT_DIR)
	protoc -I $(PROTO_DIR) $(PROTO_DIR)/*.proto \
		--plugin=protoc-gen-go=$(LOCAL_BIN)/protoc-gen-go \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--plugin=protoc-gen-go-grpc=$(LOCAL_BIN)/protoc-gen-go-grpc \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		--plugin=protoc-gen-validate=$(LOCAL_BIN)/protoc-gen-validate \
		--validate_out="lang=go,paths=source_relative:$(OUT_DIR)"

## migrate-create: Create a new migration
.PHONY: migrate-create
migrate-create: $(LOCAL_BIN)/migrate
	@test -n "$(NAME)" || (echo "Error: NAME variable is required. Usage: make migrate-create NAME=migration_name"; exit 1)
	$(LOCAL_BIN)/migrate create -ext sql -dir ${CURDIR}/migrations -seq $(NAME)

## migrate-up: Apply all pending migrations
.PHONY: migrate-up
migrate-up: $(LOCAL_BIN)/migrate
	$(LOCAL_BIN)/migrate -path=${CURDIR}/migrations -database=$(DATABASE_URL) up

## migrate-down: Rollback migrations
.PHONY: migrate-down
migrate-down: $(LOCAL_BIN)/migrate
	@read -p "Number of migrations to rollback (default: 1): " NUM; NUM=$${NUM:-1}; \
	$(LOCAL_BIN)/migrate -path=${CURDIR}/migrations -database=$(DATABASE_URL) down $$NUM

## migrate-version: Show current migration version
.PHONY: migrate-version
migrate-version: $(LOCAL_BIN)/migrate
	$(LOCAL_BIN)/migrate -path=${CURDIR}/migrations -database=$(DATABASE_URL) version

## bin-deps: Install all binary dependencies
.PHONY: bin-deps
bin-deps: \
	$(LOCAL_BIN)/golangci-lint \
	$(LOCAL_BIN)/swag \
	$(LOCAL_BIN)/migrate \
	$(LOCAL_BIN)/mockgen \
	$(LOCAL_BIN)/protoc-gen-go \
	$(LOCAL_BIN)/protoc-gen-go-grpc \
	$(LOCAL_BIN)/protoc-gen-validate

$(LOCAL_BIN)/golangci-lint:
	@echo ">>> Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@mkdir -p $(LOCAL_BIN)
	@tmp_dir=$$(mktemp -d); \
	git clone --depth 1 --branch $(GOLANGCI_LINT_VERSION) https://github.com/golangci/golangci-lint.git $$tmp_dir; \
	cd $$tmp_dir/cmd/golangci-lint; \
	go build -o $(LOCAL_BIN)/golangci-lint .; \
	cd -; \
	rm -rf $$tmp_dir
	@echo ">>> golangci-lint installed successfully"

$(LOCAL_BIN)/swag:
	@echo ">>> Installing swag $(SWAG_VERSION)..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)

$(LOCAL_BIN)/migrate:
	@echo ">>> Installing migrate $(MIGRATE_VERSION)..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

$(LOCAL_BIN)/mockgen:
	@echo ">>> Installing mockgen $(MOCKGEN_VERSION)..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)

$(LOCAL_BIN)/protoc-gen-go:
	@echo ">>> Installing protoc-gen-go $(PROTOC_GEN_GO_VERSION)..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)

$(LOCAL_BIN)/protoc-gen-go-grpc:
	@echo ">>> Installing protoc-gen-go-grpc $(PROTOC_GEN_GO_GRPC_VERSION)..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

$(LOCAL_BIN)/protoc-gen-validate:
	@echo ">>> Installing protoc-gen-validate $(PROTOC_GEN_VALIDATE_VERSION)..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@$(PROTOC_GEN_VALIDATE_VERSION)

## clean: Remove build artifacts
.PHONY: clean
clean:
	rm -rf build/

## clean-bin: Remove all installed binary tools
.PHONY: clean-bin
clean-bin:
	rm -rf $(LOCAL_BIN)

## clean-all: Remove build artifacts and binary tools
.PHONY: clean-all
clean-all: clean clean-bin


