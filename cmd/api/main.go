package main

import (
	"context"

	"github.com/gabrielbruno7/ocr-service/internal/document/handler"
	"github.com/gabrielbruno7/ocr-service/internal/document/infrainstructure"
	"github.com/gabrielbruno7/ocr-service/internal/document/usecase"
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

	queue, err := queue.New(cfg.AMQPURL)
	if err != nil {
		log.Error("erro ao conectar no RabbitMQ", "error", err)
		return
	}
	defer queue.Close()

	documentRepository := infrainstructure.NewDocumentRepository(db)

	extractDocumentUsecase := usecase.NewExtractDocumentUsecase(ctx, queue, cfg.UploadDir, documentRepository)
	getDocumentUsecase := usecase.NewGetDocumentUsecase(ctx, documentRepository)

	documentHandler := handler.NewDocumentHandler(log, getDocumentUsecase, extractDocumentUsecase)

	router := routes.New(documentHandler)

	log.Info("servidor iniciado", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Error("erro ao iniciar servidor", "error", err)
	}
}
