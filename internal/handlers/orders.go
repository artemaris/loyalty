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

func (h *OrdersHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

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

	if !h.luhnService.Validate(orderNumber) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()

	existingOrder, err := h.storage.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "Order already uploaded by another user", http.StatusConflict)
		return
	}

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

	go h.processOrderAsync(orderNumber)

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrdersHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
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

	orders, err := h.storage.GetUserOrders(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *OrdersHandler) processOrderAsync(orderNumber string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	accrualResp, err := h.accrualService.GetOrderInfo(ctx, orderNumber)
	if err != nil {
		return
	}

	if accrualResp == nil {
		return
	}

	var accrual *float64
	if accrualResp.Accrual != nil {
		accrual = accrualResp.Accrual
	}

	status := models.OrderStatus(accrualResp.Status)
	if err := h.storage.UpdateOrderStatus(ctx, orderNumber, status, accrual); err != nil {
		return
	}

	if status == models.OrderStatusProcessed && accrual != nil && *accrual > 0 {
		order, err := h.storage.GetOrderByNumber(ctx, orderNumber)
		if err != nil || order == nil {
			return
		}

		balance, err := h.storage.GetUserBalance(ctx, order.UserID)
		if err != nil {
			return
		}

		newCurrent := balance.Current + *accrual
		if err := h.storage.UpdateUserBalance(ctx, order.UserID, newCurrent, balance.Withdrawn); err != nil {
			return
		}
	}
}
