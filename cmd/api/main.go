package main

import (
	"context"
	"net/http"

	"github.com/gabrielbruno7/ocr-service/internal/document"
	"github.com/gabrielbruno7/ocr-service/internal/ocr"
	"github.com/gabrielbruno7/ocr-service/internal/platform/config"
	"github.com/gabrielbruno7/ocr-service/internal/platform/database"
	"github.com/gabrielbruno7/ocr-service/internal/platform/logger"
	"github.com/gabrielbruno7/ocr-service/internal/platform/routes"
)

func main() {
	log := logger.New()
	cfg := config.Load()

	ctx := context.Background()
	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("erro ao conectar no banco", "error", err)
		return
	}
	defer db.Close()

	processor := ocr.NewTesseractProcessor(cfg.OCRLanguage, cfg.OCRPoolSize)

	repository := document.NewRepository(db)
	docHandler := document.NewHandler(processor, repository, log)

	router := routes.New(docHandler, log)

	log.Info("servidor iniciado", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Error("erro ao iniciar servidor", "error", err)
	}
}
