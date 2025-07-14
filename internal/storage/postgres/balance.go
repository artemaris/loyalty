package postgres

import (
	"context"
	"database/sql"

	"github.com/artemaris/loyalty/internal/models"
)

func (p *PostgresStorage) GetUserBalance(ctx context.Context, userID int64) (*models.Balance, error) {
	var balance models.Balance
	query := `SELECT user_id, current, withdrawn FROM user_balances WHERE user_id = $1`

	err := p.db.GetContext(ctx, &balance, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если баланс не найден, создаем новый с нулевыми значениями
			balance = models.Balance{
				UserID:    userID,
				Current:   0,
				Withdrawn: 0,
			}
			return &balance, nil
		}
		return nil, err
	}

	return &balance, nil
}

func (p *PostgresStorage) UpdateUserBalance(ctx context.Context, userID int64, current, withdrawn float64) error {
	query := `
		INSERT INTO user_balances (user_id, current, withdrawn, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			current = $2,
			withdrawn = $3,
			updated_at = NOW()`

	_, err := p.db.ExecContext(ctx, query, userID, current, withdrawn)
	return err
}
