package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/artemaris/loyalty/internal/middleware"
	"github.com/artemaris/loyalty/internal/models"
	"github.com/artemaris/loyalty/internal/services"
	"github.com/artemaris/loyalty/internal/storage"
)

type WithdrawalsHandler struct {
	storage     storage.Storage
	luhnService *services.LuhnService
}

func NewWithdrawalsHandler(storage storage.Storage, luhnService *services.LuhnService) *WithdrawalsHandler {
	return &WithdrawalsHandler{
		storage:     storage,
		luhnService: luhnService,
	}
}

func (h *WithdrawalsHandler) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var withdrawalReq models.WithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&withdrawalReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !h.luhnService.Validate(withdrawalReq.Order) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	if withdrawalReq.Sum <= 0 {
		http.Error(w, "Sum must be positive", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	balance, err := h.storage.GetUserBalance(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if balance.Current < withdrawalReq.Sum {
		http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
		return
	}

	withdrawal := &models.Withdrawal{
		UserID:      userID,
		Order:       withdrawalReq.Order,
		Sum:         withdrawalReq.Sum,
		ProcessedAt: time.Now(),
	}

	if err := h.storage.CreateWithdrawal(ctx, withdrawal); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	newCurrent := balance.Current - withdrawalReq.Sum
	newWithdrawn := balance.Withdrawn + withdrawalReq.Sum

	if err := h.storage.UpdateUserBalance(ctx, userID, newCurrent, newWithdrawn); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WithdrawalsHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	withdrawals, err := h.storage.GetUserWithdrawals(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
