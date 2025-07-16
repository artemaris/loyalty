package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AccrualService struct {
	baseURL    string
	httpClient *http.Client
}

type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

func NewAccrualService(baseURL string) *AccrualService {
	return &AccrualService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (a *AccrualService) GetOrderInfo(ctx context.Context, orderNumber string) (*AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", a.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &accrualResp, nil

	case http.StatusNoContent:
		return nil, nil

	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("rate limit exceeded")

	case http.StatusInternalServerError:
		return nil, fmt.Errorf("internal server error from accrual system")

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
