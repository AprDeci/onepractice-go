package main

import (
	"log/slog"
	"os"

	"onepractice-golang/internal/app"
)

// @title Onepractice API
// @version 0.1.0
// @description Onepractice 在线英语真题平台 Go 后端 API。
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	application, err := app.New()
	if err != nil {
		slog.Error("initialize application", "error", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		slog.Error("run server", "error", err)
		os.Exit(1)
	}
}
