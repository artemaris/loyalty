# Настройка базы данных

## Требования

- PostgreSQL 12+
- psql (PostgreSQL client)

## Создание базы данных

1. Создайте базу данных:
```sql
CREATE DATABASE loyalty;
```

2. Создайте пользователя (опционально):
```sql
CREATE USER loyalty_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE loyalty TO loyalty_user;
```

## Выполнение миграций

**Обязательно выполните миграции перед первым запуском приложения!**

### Go-версия (рекомендуется):
```bash
go run ./cmd/migrate -uri="postgres://username:password@localhost:5432/loyalty?sslmode=disable"
```

### Через Makefile:
```bash
make migrate DATABASE_URI="postgres://username:password@localhost:5432/loyalty?sslmode=disable"
```

### Ручное выполнение SQL:
```bash
psql "postgres://username:password@localhost:5432/loyalty?sslmode=disable" -f migrations/001_init_schema.sql
```

## Структура базы данных

### Таблица `users`
- `id` - уникальный идентификатор пользователя
- `login` - уникальный логин пользователя
- `password_hash` - хеш пароля
- `created_at` - дата создания

### Таблица `orders`
- `id` - уникальный идентификатор заказа
- `number` - номер заказа (уникальный)
- `user_id` - ID пользователя (внешний ключ)
- `status` - статус обработки (NEW, PROCESSING, INVALID, PROCESSED)
- `accrual` - начисленные баллы
- `uploaded_at` - дата загрузки

### Таблица `user_balances`
- `user_id` - ID пользователя (внешний ключ)
- `current` - текущий баланс
- `withdrawn` - сумма списанных баллов
- `updated_at` - дата обновления

### Таблица `withdrawals`
- `id` - уникальный идентификатор списания
- `user_id` - ID пользователя (внешний ключ)
- `order_number` - номер заказа для списания
- `sum` - сумма списания
- `processed_at` - дата обработки

## Конфигурация

Укажите строку подключения к базе данных через:
- Переменную окружения: `DATABASE_URI`
- Флаг командной строки: `-d`

Пример:
```bash
export DATABASE_URI="postgres://loyalty_user:password@localhost:5432/loyalty?sslmode=disable"
``` 