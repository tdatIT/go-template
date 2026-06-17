APP_NAME ?= go-service
MAIN_PATH ?= ./cmd/main.go
BINARY_DIR ?= build
BINARY_NAME ?= app-bin
DOCKER_IMAGE ?= go-service:local

.PHONY: help run build test lint fmt tidy docker-build clean mocks

help:
	@echo Available targets:
	@echo   make run          - Run the service locally
	@echo   make build        - Build binary into $(BINARY_DIR)/$(BINARY_NAME)
	@echo   make test         - Run all tests
	@echo   make lint         - Run go vet
	@echo   make fmt          - Run go fmt on all packages
	@echo   make tidy         - Run go mod tidy
	@echo   make mocks        - Regenerate mocks via mockery (.mockery.yaml)
	@echo   make docker-build - Build Docker image $(DOCKER_IMAGE)
	@echo   make clean        - Remove build artifacts

run:
	go run $(MAIN_PATH)

build:
	@if not exist $(BINARY_DIR) mkdir $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PATH)

test:
	go test ./...

lint:
	go vet ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

mocks:
	go tool mockery

docker-build:
	docker build -t $(DOCKER_IMAGE) .

clean:
	@if exist $(BINARY_DIR) rmdir /s /q $(BINARY_DIR)

