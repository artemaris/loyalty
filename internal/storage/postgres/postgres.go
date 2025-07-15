package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

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

	return &PostgresStorage{db: db}, nil
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}
