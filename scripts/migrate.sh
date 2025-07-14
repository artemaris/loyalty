#!/bin/bash

# Скрипт для выполнения миграций базы данных
# Использование: ./scripts/migrate.sh <DATABASE_URI>

if [ $# -eq 0 ]; then
    echo "Usage: $0 <DATABASE_URI>"
    echo "Example: $0 'postgres://user:password@localhost:5432/loyalty?sslmode=disable'"
    exit 1
fi

DATABASE_URI=$1

echo "Running migrations on database..."

# Выполняем миграции
psql "$DATABASE_URI" -f migrations/001_init_schema.sql

if [ $? -eq 0 ]; then
    echo "Migrations completed successfully!"
else
    echo "Migration failed!"
    exit 1
fi 