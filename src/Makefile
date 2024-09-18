BIN_SRV := "./bin/rate-limiter-server"
BIN_CLI := "./bin/rate-limiter-cli"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN_SRV) -ldflags "$(LDFLAGS)" ./cmd/server
	go build -v -o $(BIN_CLI) -ldflags "$(LDFLAGS)" ./cmd/cli

run: build
	$(BIN_SRV) -config config.yaml

build-img:
	docker compose build

up: build-img
	docker compose up -d

down:
	docker compose down

version: build
	$(BIN_SRV) version

test:
	go test -race -count 100 ./internal/api/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.59.1

lint: install-lint-deps
	golangci-lint run ./...

lint-fix:
	go mod tidy
	gofumpt -l -w .
	golangci-lint run --fix

PROTO_DIR = ./proto
PROTO_FILES = $(PROTO_DIR)/login_info.proto
GENERATED_DIR = ./proto

generate:
	protoc --go_out=$(GENERATED_DIR) --go-grpc_out=$(GENERATED_DIR) $(PROTO_FILES)

mockgen:
	mockgen -source=app/lib/storage/storage.go -destination=app/lib/storage/mock_storage.go -package=storage

integration-tests:
	# Run the integration tests inside a Docker container
	docker compose -f docker-compose.integration_test.yml up --build --abort-on-container-exit \

	# Clean up the environment
	docker compose -f docker-compose.integration_test.yml down

DOCKER_REGISTRY := docker.io
DOCKER_USER := thejubadze
DOCKER_PATH := $(DOCKER_REGISTRY)/$(DOCKER_USER)
push: build-img
	docker push $(DOCKER_PATH)/rate-limiter:latest

DB_URL := postgres://root:123@localhost:5432/rate-limiter?sslmode=disable
GOOSE_BIN := $(GOPATH)/bin/goose
MIGRATIONS_DIR := migrations

goose-up:
	$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

goose-down:
	$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

goose-status:
	$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

goose-create:
	$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) create $(name) sql

.PHONY: build run build-img up down version test install-lint-deps lint lint-fix generate mockgen integration-tests push goose-up goose-down goose-status goose-create
