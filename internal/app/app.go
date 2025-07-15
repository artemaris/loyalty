package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/artemaris/loyalty/internal/config"
	"github.com/artemaris/loyalty/internal/handlers"
	"github.com/artemaris/loyalty/internal/middleware"
	"github.com/artemaris/loyalty/internal/services"
	"github.com/artemaris/loyalty/internal/storage"
	postgres "github.com/artemaris/loyalty/internal/storage/postgres"
)

type App struct {
	cfg                *config.Config
	server             *http.Server
	storage            storage.Storage
	authService        *services.AuthService
	authHandler        *handlers.AuthHandler
	ordersHandler      *handlers.OrdersHandler
	balanceHandler     *handlers.BalanceHandler
	withdrawalsHandler *handlers.WithdrawalsHandler
}

func New(cfg *config.Config) (*App, error) {
	log.Printf("Initializing application with config: RunAddress=%s, DatabaseURI=%s, AccrualSystemAddress=%s",
		cfg.RunAddress, cfg.DatabaseURI, cfg.AccrualSystemAddress)

	log.Println("Connecting to database...")
	storage, err := postgres.New(cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	log.Println("Database connection established successfully")

	log.Println("Initializing services...")
	authService := services.NewAuthService(cfg.JWTSecret)
	luhnService := services.NewLuhnService()
	accrualService := services.NewAccrualService(cfg.AccrualSystemAddress)

	log.Println("Initializing handlers...")
	authHandler := handlers.NewAuthHandler(storage, authService)
	ordersHandler := handlers.NewOrdersHandler(storage, luhnService, accrualService)
	balanceHandler := handlers.NewBalanceHandler(storage)
	withdrawalsHandler := handlers.NewWithdrawalsHandler(storage, luhnService)

	app := &App{
		cfg:                cfg,
		storage:            storage,
		authService:        authService,
		authHandler:        authHandler,
		ordersHandler:      ordersHandler,
		balanceHandler:     balanceHandler,
		withdrawalsHandler: withdrawalsHandler,
	}

	app.server = &http.Server{
		Addr:    cfg.RunAddress,
		Handler: app.setupRoutes(),
	}

	log.Println("Application initialized successfully")
	return app, nil
}

func (a *App) Run(ctx context.Context) error {
	log.Printf("Starting server on %s", a.cfg.RunAddress)

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	if err := a.storage.Close(); err != nil {
		log.Printf("Failed to close storage: %v", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}

func (a *App) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	authMiddleware := middleware.NewAuthMiddleware(a.authService)

	mux.HandleFunc("/api/user/register", a.authHandler.Register)
	mux.HandleFunc("/api/user/login", a.authHandler.Login)

	mux.Handle("/api/user/orders", authMiddleware.Authenticate(http.HandlerFunc(a.handleOrders)))
	mux.Handle("/api/user/balance", authMiddleware.Authenticate(http.HandlerFunc(a.handleBalance)))
	mux.Handle("/api/user/balance/withdraw", authMiddleware.Authenticate(http.HandlerFunc(a.handleWithdrawals)))
	mux.Handle("/api/user/withdrawals", authMiddleware.Authenticate(http.HandlerFunc(a.handleGetWithdrawals)))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("GopherMart API"))
		if err != nil {
			println(err)
		}
	})

	return middleware.Compression(mux)
}

func (a *App) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.ordersHandler.UploadOrder(w, r)
	case http.MethodGet:
		a.ordersHandler.GetOrders(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *App) handleBalance(w http.ResponseWriter, r *http.Request) {
	a.balanceHandler.GetBalance(w, r)
}

func (a *App) handleWithdrawals(w http.ResponseWriter, r *http.Request) {
	a.withdrawalsHandler.CreateWithdrawal(w, r)
}

func (a *App) handleGetWithdrawals(w http.ResponseWriter, r *http.Request) {
	a.withdrawalsHandler.GetWithdrawals(w, r)
}
