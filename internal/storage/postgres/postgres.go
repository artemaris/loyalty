package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/artemaris/loyalty/internal/storage"
)

type PostgresStorage struct {
	db *sqlx.DB
}

// New создает новое подключение к PostgreSQL
func New(databaseURI string) (storage.Storage, error) {
	db, err := sqlx.Connect("postgres", databaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Создаем таблицы, если они не существуют
	if err := createTablesIfNotExist(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}

// createTablesIfNotExist создает таблицы, если они не существуют
func createTablesIfNotExist(db *sqlx.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			login VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id BIGSERIAL PRIMARY KEY,
			number VARCHAR(255) UNIQUE NOT NULL,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			status VARCHAR(50) NOT NULL DEFAULT 'NEW',
			accrual DECIMAL(10,2),
			uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS user_balances (
			user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			current DECIMAL(10,2) NOT NULL DEFAULT 0,
			withdrawn DECIMAL(10,2) NOT NULL DEFAULT 0,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS withdrawals (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			order_number VARCHAR(255) NOT NULL,
			sum DECIMAL(10,2) NOT NULL,
			processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_uploaded_at ON orders(uploaded_at)`,
		`CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_withdrawals_processed_at ON withdrawals(processed_at)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query '%s': %w", query, err)
		}
	}

	return nil
}
