package main

import (
	"context"
	"log"
	"net/http"

	"energy-platform/internal/config"
	"energy-platform/internal/db"
	"energy-platform/internal/seed"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	pool, err := db.Connect(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Print("migrations applied")

	if err := seed.Run(ctx, pool, cfg.DataDir); err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Print("seed complete")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
