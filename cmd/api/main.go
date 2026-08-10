package main

import (
	"context"

	"github.com/gabrielbruno7/ocr-service/internal/document"
	"github.com/gabrielbruno7/ocr-service/internal/platform/config"
	"github.com/gabrielbruno7/ocr-service/internal/platform/database"
	"github.com/gabrielbruno7/ocr-service/internal/platform/logger"
	"github.com/gabrielbruno7/ocr-service/internal/platform/queue"
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

	q, err := queue.New(cfg.AMQPURL)
	if err != nil {
		log.Error("erro ao conectar no RabbitMQ", "error", err)
		return
	}
	defer q.Close()

	repository := document.NewRepository(db)
	docHandler := document.NewHandler(repository, q, cfg.UploadDir, log)

	router := routes.New(docHandler)

	log.Info("servidor iniciado", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Error("erro ao iniciar servidor", "error", err)
	}
}
