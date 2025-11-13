.PHONY: build install clean test fmt lint run help

BINARY_NAME=gomm
INSTALL_PATH=$(HOME)/go/bin

help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Собрать проект
	@echo "Сборка $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) ./cmd/gomm
	@echo "Сборка завершена: bin/$(BINARY_NAME)"

install: build ## Установить в $HOME/go/bin
	@echo "Установка $(BINARY_NAME) в $(INSTALL_PATH)..."
	cp bin/$(BINARY_NAME) $(INSTALL_PATH)/
	@echo "Установлено: $(INSTALL_PATH)/$(BINARY_NAME)"

clean: ## Удалить артефакты сборки
	@echo "Очистка..."
	rm -rf bin/
	go clean
	@echo "Очистка завершена"

test: ## Запустить unit тесты
	@echo "Запуск unit тестов..."
	go test -v -race ./internal/... ./pkg/...

test-e2e: build ## Запустить E2E тесты
	@echo "Запуск E2E тестов..."
	@echo "Сборка бинарника для тестов..."
	@mkdir -p bin
	@GOMM_BINARY=$(PWD)/bin/$(BINARY_NAME) go test -v -timeout 30m ./tests/e2e/...

test-e2e-init: build ## Запустить тесты инициализации
	@echo "Запуск тестов инициализации..."
	@GOMM_BINARY=$(PWD)/bin/$(BINARY_NAME) go test -v -timeout 10m ./tests/e2e/ -run TestInit

test-e2e-clone: build ## Запустить тесты клонирования
	@echo "Запуск тестов клонирования..."
	@GOMM_BINARY=$(PWD)/bin/$(BINARY_NAME) go test -v -timeout 10m ./tests/e2e/ -run TestClone

test-e2e-workspace: build ## Запустить тесты workspace
	@echo "Запуск тестов workspace..."
	@GOMM_BINARY=$(PWD)/bin/$(BINARY_NAME) go test -v -timeout 10m ./tests/e2e/ -run "Test.*[Ww]orkspace|Test.*[Ll]ocal|Test.*[Rr]emote|Test.*[Rr]eplace"

test-e2e-versioning: build ## Запустить тесты версионирования
	@echo "Запуск тестов версионирования..."
	@GOMM_BINARY=$(PWD)/bin/$(BINARY_NAME) go test -v -timeout 10m ./tests/e2e/ -run TestVersion

test-e2e-release: build ## Запустить тесты релизов
	@echo "Запуск тестов релизов..."
	@GOMM_BINARY=$(PWD)/bin/$(BINARY_NAME) go test -v -timeout 10m ./tests/e2e/ -run TestRelease

test-e2e-workflow: build ## Запустить полные workflow тесты
	@echo "Запуск полных workflow тестов..."
	@GOMM_BINARY=$(PWD)/bin/$(BINARY_NAME) go test -v -timeout 20m ./tests/e2e/ -run "Test.*[Ww]orkflow"

test-all: test test-e2e ## Запустить все тесты

test-coverage: ## Запустить unit тесты с coverage
	@echo "Запуск тестов с coverage..."
	go test -v -race -coverprofile=coverage.out ./internal/... ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

fmt: ## Форматирование кода
	@echo "Форматирование кода..."
	go fmt ./...
	@echo "Форматирование завершено"

lint: ## Проверка линтером
	@echo "Проверка кода линтером..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint не установлен. Установите: https://golangci-lint.run/usage/install/"; \
	fi

run: build ## Собрать и запустить
	./bin/$(BINARY_NAME)

dev-deps: ## Установить зависимости для разработки
	@echo "Установка зависимостей для разработки..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Зависимости установлены"

mod-tidy: ## Запустить go mod tidy
	@echo "Обновление зависимостей..."
	go mod tidy
	@echo "Зависимости обновлены"

mod-download: ## Скачать зависимости
	@echo "Скачивание зависимостей..."
	go mod download
	@echo "Зависимости скачаны"

.DEFAULT_GOAL := help
