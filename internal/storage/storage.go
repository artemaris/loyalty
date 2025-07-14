package storage

import (
	"context"

	"github.com/artemaris/loyalty/internal/models"
)

// Storage интерфейс для работы с базой данных
type Storage interface {
	// Пользователи
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)

	// Заказы
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrderByNumber(ctx context.Context, number string) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID int64) ([]models.Order, error)
	UpdateOrderStatus(ctx context.Context, number string, status models.OrderStatus, accrual *float64) error

	// Баланс
	GetUserBalance(ctx context.Context, userID int64) (*models.Balance, error)
	UpdateUserBalance(ctx context.Context, userID int64, current, withdrawn float64) error

	// Списания
	CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error
	GetUserWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error)

	// Закрытие соединения
	Close() error
}
