package main

import (
	"context"
	"github.com/artemaris/loyalty/internal/accrual"
	httpapi "github.com/artemaris/loyalty/internal/http"
	"github.com/artemaris/loyalty/internal/storage"
	"log"
	"net/http"
	"os"
)

func main() {
	dsn := os.Getenv("DATABASE_URI")
	store, err := storage.New(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accrualAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	accrual.StartWorker(ctx, store, accrualAddress)

	handler := httpapi.NewRouter(store)
	addr := os.Getenv("RUN_ADDRESS")
	log.Printf("starting server at %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
