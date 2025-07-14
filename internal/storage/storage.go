package storage

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	DB *pgxpool.Pool
}

func New(connString string) (*Storage, error) {
	db, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	return &Storage{DB: db}, nil
}

func (s *Storage) Close() {
	s.DB.Close()
}
