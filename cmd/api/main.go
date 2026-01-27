package main

import (
	"context"
	"log"
	"net/http"

	"github.com/<your-username>/txnflow/internal/config"
	"github.com/<your-username>/txnflow/internal/db"
	httpapi "github.com/<your-username>/txnflow/internal/http"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	handlers := httpapi.NewHandlers(pool)
	router := httpapi.Router(handlers)

	log.Printf("API listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
