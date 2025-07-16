package postgres

import (
	"context"
	"database/sql"

	"github.com/artemaris/loyalty/internal/models"
)

func (p *PostgresStorage) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (login, password_hash, created_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	return p.db.GetContext(ctx, user, query, user.Login, user.Password, user.Created)
}

func (p *PostgresStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`

	err := p.db.GetContext(ctx, &user, query, login)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (p *PostgresStorage) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	query := `SELECT id, login, password_hash, created_at FROM users WHERE id = $1`

	err := p.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
