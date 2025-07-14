package auth

import (
	"context"
	"errors"
	"github.com/artemaris/loyalty/internal/storage"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

func Register(ctx context.Context, s *storage.Storage, login, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(ctx, "INSERT INTO users (login, password_hash) VALUES ($1, $2)", login, string(hash))
	return err
}

func Authenticate(ctx context.Context, s *storage.Storage, login, password string) (int, error) {
	var (
		id   int
		hash string
	)
	err := s.DB.QueryRow(ctx, "SELECT id, password_hash FROM users WHERE login=$1", login).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return 0, ErrInvalidCredentials
	}
	return id, nil
}
