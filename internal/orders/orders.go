package orders

import (
	"context"
	"errors"
	"github.com/artemaris/loyalty/internal/storage"
)

var ErrInvalidOrder = errors.New("invalid order number")

func ValidateLuhn(number string) bool {
	var sum int
	n := len(number)
	alt := false
	for i := n - 1; i >= 0; i-- {
		num := int(number[i] - '0')
		if num < 0 || num > 9 {
			return false
		}
		if alt {
			num *= 2
			if num > 9 {
				num -= 9
			}
		}
		sum += num
		alt = !alt
	}
	return sum%10 == 0
}

func CreateOrder(ctx context.Context, s *storage.Storage, userID int, number string) error {
	if !ValidateLuhn(number) {
		return ErrInvalidOrder
	}
	_, err := s.DB.Exec(ctx, `INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3)`, userID, number, "NEW")
	return err
}
