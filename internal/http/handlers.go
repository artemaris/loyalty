package http

import (
	"database/sql"
	"encoding/json"
	"github.com/artemaris/loyalty/internal/auth"
	"github.com/artemaris/loyalty/internal/orders"
	"github.com/artemaris/loyalty/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"net/http"
)

type Server struct {
	store *storage.Storage
}

func NewRouter(store *storage.Storage) http.Handler {
	s := &Server{store: store}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))

	r.Post("/api/user/register", s.Register)
	r.Post("/api/user/login", s.Login)

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)
		r.Post("/api/user/orders", s.UploadOrder)
		r.Get("/api/user/orders", s.GetOrders)
		r.Get("/api/user/balance", s.GetBalance)
		r.Post("/api/user/balance/withdraw", s.Withdraw)
		r.Get("/api/user/withdrawals", s.GetWithdrawals)
	})

	return r
}

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	err := auth.Register(r.Context(), s.store, body.Login, body.Password)
	if err != nil {
		http.Error(w, "user exists", http.StatusConflict)
		return
	}
	id, _ := auth.Authenticate(r.Context(), s.store, body.Login, body.Password)
	token, _ := auth.GenerateToken(id)
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	id, err := auth.Authenticate(r.Context(), s.store, body.Login, body.Password)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	token, _ := auth.GenerateToken(id)
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	number := string(body)
	err = orders.CreateOrder(r.Context(), s.store, userID, number)
	if err != nil {
		http.Error(w, "unprocessable", http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	var balance, withdrawn float64
	err := s.store.DB.QueryRow(r.Context(), `
		SELECT COALESCE(SUM(accrual), 0), (
			SELECT COALESCE(SUM(amount), 0) FROM withdrawals WHERE user_id=$1
		)
		FROM orders WHERE user_id=$1 AND status='PROCESSED'
	`, userID).Scan(&balance, &withdrawn)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"current":   balance - withdrawn,
		"withdrawn": withdrawn,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	type req struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var balance float64
	err := s.store.DB.QueryRow(r.Context(), `
		SELECT COALESCE(SUM(accrual), 0) - (
			SELECT COALESCE(SUM(amount), 0) FROM withdrawals WHERE user_id=$1
		)
		FROM orders WHERE user_id=$1 AND status='PROCESSED'
	`, userID).Scan(&balance)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if balance < body.Sum {
		http.Error(w, "not enough balance", http.StatusPaymentRequired)
		return
	}

	_, err = s.store.DB.Exec(r.Context(), `
		INSERT INTO withdrawals (user_id, order_number, amount) VALUES ($1, $2, $3)
	`, userID, body.Order, body.Sum)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	rows, err := s.store.DB.Query(r.Context(), `
		SELECT order_number, amount, processed_at 
		FROM withdrawals 
		WHERE user_id=$1 ORDER BY processed_at DESC
	`, userID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type withdrawal struct {
		Order       string  `json:"order"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}
	var result []withdrawal

	for rows.Next() {
		var wItem withdrawal
		if err := rows.Scan(&wItem.Order, &wItem.Sum, &wItem.ProcessedAt); err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		result = append(result, wItem)
	}

	if len(result) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	rows, err := s.store.DB.Query(r.Context(), `
		SELECT number, status, accrual, uploaded_at 
		FROM orders 
		WHERE user_id=$1 ORDER BY uploaded_at DESC`, userID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type order struct {
		Number     string   `json:"number"`
		Status     string   `json:"status"`
		Accrual    *float64 `json:"accrual,omitempty"`
		UploadedAt string   `json:"uploaded_at"`
	}

	var result []order
	for rows.Next() {
		var o order
		var accrual sql.NullFloat64
		if err := rows.Scan(&o.Number, &o.Status, &accrual, &o.UploadedAt); err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		if accrual.Valid {
			o.Accrual = &accrual.Float64
		}
		result = append(result, o)
	}
	if len(result) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
