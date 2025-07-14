package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/artemaris/loyalty/internal/storage"
	"net/http"
	"time"
)

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func StartWorker(ctx context.Context, s *storage.Storage, accrualAddress string) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				updateOrders(ctx, s, accrualAddress)
			}
		}
	}()
}

func updateOrders(ctx context.Context, s *storage.Storage, accrualAddress string) {
	rows, err := s.DB.Query(ctx, "SELECT id, number FROM orders WHERE status IN ('NEW', 'PROCESSING')")
	if err != nil {
		fmt.Println("error querying orders:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id     int
			number string
		)
		if err := rows.Scan(&id, &number); err != nil {
			continue
		}
		status, accrual, err := fetchAccrual(ctx, accrualAddress, number)
		if err != nil {
			continue
		}
		_, err = s.DB.Exec(ctx, "UPDATE orders SET status=$1, accrual=$2 WHERE id=$3", status, accrual, id)
		if err != nil {
			fmt.Println("error updating order:", err)
		}
	}
}

func fetchAccrual(ctx context.Context, addr, number string) (string, float64, error) {
	url := fmt.Sprintf("%s/api/orders/%s", addr, number)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var ar AccrualResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return "", 0, err
	}
	return ar.Status, ar.Accrual, nil
}
