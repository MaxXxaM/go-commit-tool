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

test: ## Запустить тесты
	@echo "Запуск тестов..."
	go test -v -race ./...

test-coverage: ## Запустить тесты с coverage
	@echo "Запуск тестов с coverage..."
	go test -v -race -coverprofile=coverage.out ./...
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
