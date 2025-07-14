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

// CreateWithdrawal обрабатывает запрос на списание баллов
func (h *WithdrawalsHandler) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID пользователя из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Парсим JSON запрос
	var withdrawalReq models.WithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&withdrawalReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверяем номер заказа с помощью алгоритма Луна
	if !h.luhnService.Validate(withdrawalReq.Order) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// Проверяем сумму списания
	if withdrawalReq.Sum <= 0 {
		http.Error(w, "Sum must be positive", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Получаем текущий баланс пользователя
	balance, err := h.storage.GetUserBalance(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Проверяем, достаточно ли средств
	if balance.Current < withdrawalReq.Sum {
		http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
		return
	}

	// Создаем списание
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

	// Обновляем баланс пользователя
	newCurrent := balance.Current - withdrawalReq.Sum
	newWithdrawn := balance.Withdrawn + withdrawalReq.Sum

	if err := h.storage.UpdateUserBalance(ctx, userID, newCurrent, newWithdrawn); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals возвращает историю списаний пользователя
func (h *WithdrawalsHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID пользователя из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	// Получаем списания пользователя
	withdrawals, err := h.storage.GetUserWithdrawals(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Отправляем JSON ответ
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
