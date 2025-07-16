package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/artemaris/loyalty/internal/middleware"
	"github.com/artemaris/loyalty/internal/storage"
)

type BalanceHandler struct {
	storage storage.Storage
}

func NewBalanceHandler(storage storage.Storage) *BalanceHandler {
	return &BalanceHandler{
		storage: storage,
	}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
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

	balance, err := h.storage.GetUserBalance(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(balance); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
