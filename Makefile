.PHONY: build test migrate test-api clean help

# Переменные
BINARY_NAME=gophermart
MIGRATE_BINARY=migrate
TEST_API_BINARY=test-api

# Сборка основного приложения
build:
	@echo "Building main application..."
	go build -o $(BINARY_NAME) ./cmd/gophermart

# Сборка утилит
build-tools:
	@echo "Building tools..."
	go build -o $(MIGRATE_BINARY) ./cmd/migrate
	go build -o $(TEST_API_BINARY) ./cmd/test-api

# Запуск тестов
test:
	@echo "Running tests..."
	go test ./... -v

# Выполнение миграций
migrate:
	@echo "Building migration tool..."
	go build -o $(MIGRATE_BINARY) ./cmd/migrate
	@echo "Running migrations..."
	./$(MIGRATE_BINARY) -uri="$(DATABASE_URI)"

# Тестирование API
test-api:
	@echo "Building API test tool..."
	go build -o $(TEST_API_BINARY) ./cmd/test-api
	@echo "Testing API..."
	./$(TEST_API_BINARY) -url="$(API_URL)"

# Полная сборка
all: build build-tools

# Очистка
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME) $(MIGRATE_BINARY) $(TEST_API_BINARY)

# Запуск приложения
run: build
	@echo "Starting application..."
	./$(BINARY_NAME)

# Помощь
help:
	@echo "Available commands:"
	@echo "  build      - Build main application"
	@echo "  build-tools- Build migration and test tools"
	@echo "  test       - Run all tests"
	@echo "  migrate    - Run database migrations (requires DATABASE_URI)"
	@echo "  test-api   - Test API endpoints (requires API_URL)"
	@echo "  all        - Build everything"
	@echo "  clean      - Clean build artifacts"
	@echo "  run        - Build and run application"
	@echo "  help       - Show this help"
	@echo ""
	@echo "Environment variables:"
	@echo "  DATABASE_URI - PostgreSQL connection string"
	@echo "  API_URL      - Base URL for API testing"
	@echo ""
	@echo "Examples:"
	@echo "  make migrate DATABASE_URI='postgres://user:pass@localhost:5432/db?sslmode=disable'"
	@echo "  make test-api API_URL='http://localhost:8080'" 