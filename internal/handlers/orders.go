package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/artemaris/loyalty/internal/middleware"
	"github.com/artemaris/loyalty/internal/models"
	"github.com/artemaris/loyalty/internal/services"
	"github.com/artemaris/loyalty/internal/storage"
)

type OrdersHandler struct {
	storage        storage.Storage
	luhnService    *services.LuhnService
	accrualService *services.AccrualService
}

func NewOrdersHandler(storage storage.Storage, luhnService *services.LuhnService, accrualService *services.AccrualService) *OrdersHandler {
	return &OrdersHandler{
		storage:        storage,
		luhnService:    luhnService,
		accrualService: accrualService,
	}
}

// UploadOrder обрабатывает загрузку номера заказа
func (h *OrdersHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
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

	// Читаем номер заказа из тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := string(body)
	if orderNumber == "" {
		http.Error(w, "Order number is required", http.StatusBadRequest)
		return
	}

	// Проверяем номер заказа с помощью алгоритма Луна
	if !h.luhnService.Validate(orderNumber) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()

	// Проверяем, существует ли заказ с таким номером
	existingOrder, err := h.storage.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existingOrder != nil {
		// Если заказ уже существует у этого пользователя
		if existingOrder.UserID == userID {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Если заказ существует у другого пользователя
		http.Error(w, "Order already uploaded by another user", http.StatusConflict)
		return
	}

	// Создаем новый заказ
	order := &models.Order{
		Number:     orderNumber,
		UserID:     userID,
		Status:     models.OrderStatusNew,
		UploadedAt: time.Now(),
	}

	if err := h.storage.CreateOrder(ctx, order); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Запускаем фоновую обработку заказа
	go h.processOrderAsync(orderNumber)

	w.WriteHeader(http.StatusAccepted)
}

// GetOrders возвращает список заказов пользователя
func (h *OrdersHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
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

	// Получаем заказы пользователя
	orders, err := h.storage.GetUserOrders(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Отправляем JSON ответ
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// processOrderAsync обрабатывает заказ асинхронно
func (h *OrdersHandler) processOrderAsync(orderNumber string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Получаем информацию о заказе из внешней системы
	accrualResp, err := h.accrualService.GetOrderInfo(ctx, orderNumber)
	if err != nil {
		// В случае ошибки оставляем статус NEW
		return
	}

	if accrualResp == nil {
		// Заказ не найден в системе начислений
		return
	}

	// Обновляем статус заказа
	var accrual *float64
	if accrualResp.Accrual != nil {
		accrual = accrualResp.Accrual
	}

	status := models.OrderStatus(accrualResp.Status)
	if err := h.storage.UpdateOrderStatus(ctx, orderNumber, status, accrual); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		return
	}

	// Если заказ обработан и есть начисление, обновляем баланс пользователя
	if status == models.OrderStatusProcessed && accrual != nil && *accrual > 0 {
		// Получаем заказ для определения пользователя
		order, err := h.storage.GetOrderByNumber(ctx, orderNumber)
		if err != nil || order == nil {
			return
		}

		// Получаем текущий баланс
		balance, err := h.storage.GetUserBalance(ctx, order.UserID)
		if err != nil {
			return
		}

		// Обновляем баланс
		newCurrent := balance.Current + *accrual
		if err := h.storage.UpdateUserBalance(ctx, order.UserID, newCurrent, balance.Withdrawn); err != nil {
			return
		}
	}
}
