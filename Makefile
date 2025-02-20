# Makefile для проекта HEPop-Go

# Переменные
BINARY_NAME=hepop-go
SRC_DIR=./cmd/hepop
CONFIG_FILE=config/config.yaml

# Команды
.PHONY: all build test clean run

all: build

build:
	@echo "Сборка проекта..."
	go build -o $(BINARY_NAME) $(SRC_DIR)

test:
	@echo "Запуск тестов..."
	go test ./... -v

clean:
	@echo "Очистка..."
	rm -f $(BINARY_NAME)

run: build
	@echo "Запуск проекта..."
	./$(BINARY_NAME) -config $(CONFIG_FILE)

lint:
	@echo "Запуск линтера..."
	golangci-lint run

fmt:
	@echo "Форматирование кода..."
	go fmt ./...

tidy:
	@echo "Обновление зависимостей..."
	go mod tidy

# Дополнительные команды
docker-build:
	@echo "Сборка Docker образа..."
	docker build -t $(BINARY_NAME):latest .

docker-run:
	@echo "Запуск Docker контейнера..."
	docker run --rm -p 8080:8080 -p 9060:9060 $(BINARY_NAME):latest

# Команда помощи
help:
	@echo "Доступные команды:"
	@echo "  make build       - Сборка проекта"
	@echo "  make test        - Запуск тестов"
	@echo "  make clean       - Очистка артефактов сборки"
	@echo "  make run         - Сборка и запуск проекта"
	@echo "  make lint        - Запуск линтера"
	@echo "  make fmt         - Форматирование кода"
	@echo "  make tidy        - Обновление зависимостей"
	@echo "  make docker-build - Сборка Docker образа"
	@echo "  make docker-run  - Запуск Docker контейнера"