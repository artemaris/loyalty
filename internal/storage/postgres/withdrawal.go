package postgres

import (
	"context"

	"github.com/artemaris/loyalty/internal/models"
)

func (p *PostgresStorage) CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error {
	query := `
		INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, processed_at`

	return p.db.GetContext(ctx, withdrawal, query,
		withdrawal.UserID, withdrawal.Order, withdrawal.Sum, withdrawal.ProcessedAt)
}

func (p *PostgresStorage) GetUserWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	var withdrawals []models.Withdrawal
	query := `
		SELECT id, user_id, order_number, sum, processed_at 
		FROM withdrawals 
		WHERE user_id = $1 
		ORDER BY processed_at DESC`

	err := p.db.SelectContext(ctx, &withdrawals, query, userID)
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}
