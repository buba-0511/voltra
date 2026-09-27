package main

import (
	"context"
	"log"
	"net/http"

	"energy-platform/internal/config"
	"energy-platform/internal/db"
	sqlcgen "energy-platform/internal/db/sqlc"
	"energy-platform/internal/httpserver"
	"energy-platform/internal/seed"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is not set - refusing to start with no signing secret")
	}

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

	queries := sqlcgen.New(pool)
	handler := httpserver.New(cfg, queries)

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatal(err)
	}
}
