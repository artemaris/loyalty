# GopherMart - Система лояльности

Система лояльности "Гофермарт" - HTTP API для управления баллами лояльности пользователей.

## Возможности

- ✅ Регистрация и аутентификация пользователей
- ✅ JWT токены для авторизации
- ✅ Загрузка номеров заказов с проверкой алгоритмом Луна
- ✅ Получение списка заказов пользователя
- ✅ Управление балансом лояльности
- ✅ Списание баллов лояльности
- ✅ История списаний пользователя
- ✅ Интеграция с внешней системой начислений
- ✅ Асинхронная обработка заказов
- ✅ Сжатие HTTP ответов (gzip)
- ✅ PostgreSQL для хранения данных

## Требования

- Go 1.24+
- PostgreSQL 12+
- psql (PostgreSQL client)

## Установка и настройка

### 1. Клонирование репозитория
```bash
git clone <repository-url>
cd loyalty
```

### 2. Установка зависимостей
```bash
go mod download
```

### 3. Настройка базы данных

Создайте базу данных PostgreSQL:
```sql
CREATE DATABASE loyalty;
```

Выполните миграции:
```bash
./scripts/migrate.sh "postgres://username:password@localhost:5432/loyalty?sslmode=disable"
```

### 4. Настройка переменных окружения

Создайте файл `.env` или установите переменные окружения:

```bash
export RUN_ADDRESS=":8080"
export DATABASE_URI="postgres://username:password@localhost:5432/loyalty?sslmode=disable"
export ACCRUAL_SYSTEM_ADDRESS="http://localhost:8081"
export JWT_SECRET="your-secret-key-here"
```

## Запуск

### Сборка приложения
```bash
go build ./cmd/gophermart
```

### Запуск сервера
```bash
./gophermart
```

Или с параметрами:
```bash
./gophermart -a :8080 -d "postgres://username:password@localhost:5432/loyalty?sslmode=disable" -r "http://localhost:8081"
```

## API Endpoints

### Публичные endpoints
- `POST /api/user/register` - Регистрация пользователя
- `POST /api/user/login` - Вход пользователя

### Защищенные endpoints (требуют JWT токен)
- `POST /api/user/orders` - Загрузка номера заказа
- `GET /api/user/orders` - Получение списка заказов
- `GET /api/user/balance` - Получение баланса
- `POST /api/user/balance/withdraw` - Списание баллов
- `GET /api/user/withdrawals` - Получение истории списаний

## Тестирование

### Запуск тестов
```bash
go test ./...
```

### Тестирование API
```bash
./scripts/test_api.sh http://localhost:8080
```

## Структура проекта

```
loyalty/
├── cmd/gophermart/     # Точка входа приложения
├── internal/
│   ├── app/           # Основная логика приложения
│   ├── config/        # Конфигурация
│   ├── handlers/      # HTTP handlers
│   ├── middleware/    # HTTP middleware
│   ├── models/        # Модели данных
│   ├── services/      # Бизнес-логика
│   └── storage/       # Работа с базой данных
├── migrations/        # SQL миграции
└── scripts/          # Вспомогательные скрипты
```

## Разработка

### Добавление новых endpoints
1. Создайте handler в `internal/handlers/`
2. Добавьте маршрут в `internal/app/app.go`
3. Добавьте тесты

### Добавление новых миграций
1. Создайте SQL файл в `migrations/`
2. Обновите скрипт `scripts/migrate.sh`

## Лицензия

MIT
