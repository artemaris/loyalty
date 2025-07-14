package postgres

import (
	"context"
	"database/sql"

	"github.com/artemaris/loyalty/internal/models"
)

func (p *PostgresStorage) CreateOrder(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO orders (number, user_id, status, accrual, uploaded_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, uploaded_at`

	return p.db.GetContext(ctx, order, query,
		order.Number, order.UserID, order.Status, order.Accrual, order.UploadedAt)
}

func (p *PostgresStorage) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	var order models.Order
	query := `SELECT id, number, user_id, status, accrual, uploaded_at FROM orders WHERE number = $1`

	err := p.db.GetContext(ctx, &order, query, number)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &order, nil
}

func (p *PostgresStorage) GetUserOrders(ctx context.Context, userID int64) ([]models.Order, error) {
	var orders []models.Order
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at 
		FROM orders 
		WHERE user_id = $1 
		ORDER BY uploaded_at DESC`

	err := p.db.SelectContext(ctx, &orders, query, userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (p *PostgresStorage) UpdateOrderStatus(ctx context.Context, number string, status models.OrderStatus, accrual *float64) error {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`
	_, err := p.db.ExecContext(ctx, query, status, accrual, number)
	return err
}
