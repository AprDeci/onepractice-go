package main

import (
	"log/slog"
	"os"

	"onepractice-golang/internal/auth"
	"onepractice-golang/internal/config"
	"onepractice-golang/internal/db"
	"onepractice-golang/internal/router"
)

func main() {
	cfg := config.Load()

	auth.Init(cfg.Auth)

	database, err := db.Open(cfg.Database)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}

	app := router.New(cfg, database)

	if err := app.Run(":" + cfg.Server.Port); err != nil {
		slog.Error("run server", "error", err)
		os.Exit(1)
	}
}
