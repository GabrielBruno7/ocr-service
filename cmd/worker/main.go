package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/ocr-service/ocr-service/internal/document/domain"
	"github.com/ocr-service/ocr-service/internal/document/infrainstructure"
	"github.com/ocr-service/ocr-service/internal/document/usecase"
	"github.com/ocr-service/ocr-service/internal/ocr"
	"github.com/ocr-service/ocr-service/internal/platform/config"
	"github.com/ocr-service/ocr-service/internal/platform/database"
	"github.com/ocr-service/ocr-service/internal/platform/logger"
	"github.com/ocr-service/ocr-service/internal/platform/queue"
	amqp "github.com/rabbitmq/amqp091-go"
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

	documentRepository := infrainstructure.NewDocumentRepository(db)
	processor := ocr.NewTesseractProcessor(cfg.OCRLanguage, cfg.OCRPoolSize)

	msgs, err := q.Consume()
	if err != nil {
		log.Error("erro ao consumir fila", "error", err)
		return
	}

	log.Info("worker iniciado, aguardando mensagens...")

	for msg := range msgs {
		handleMessage(ctx, msg, documentRepository, processor, cfg.UploadDir, log, cfg.ProcessingTimeout)
	}
}

func handleMessage(
	ctx context.Context,
	msg amqp.Delivery,
	repository domain.DocumentRepositoryInterface,
	processor ocr.Processor,
	uploadDir string,
	log *slog.Logger,
	processingTimeout time.Duration,
) {
	var job usecase.ExtractDocumentJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		log.Error("mensagem inválida, descartando", "error", err)
		msg.Nack(false, false)
		return
	}

	id := job.DocumentID

	doc, err := repository.GetByID(ctx, id)
	if err != nil {
		log.Error("erro ao buscar documento", "document_id", id, "error", err)
		msg.Nack(false, true)
		return
	}

	if doc.Status == domain.StatusDone {
		log.Info("documento já processado, ignorando", "document_id", id)
		msg.Ack(false)
		return
	}

	if doc.Status == domain.StatusProcessing {
		if doc.ProcessingStartedAt != nil && time.Since(*doc.ProcessingStartedAt) < processingTimeout {
			log.Info("documento já está sendo processado por outro worker, ignorando",
				"document_id", id, "processing_started_at", doc.ProcessingStartedAt)
			msg.Ack(false)
			return
		}
		log.Warn("documento estava em processamento há muito tempo, reprocessando",
			"document_id", id, "processing_started_at", doc.ProcessingStartedAt)
	}

	if err := repository.MarkAsProcessing(ctx, id); err != nil {
		log.Error("erro ao marcar como processing", "document_id", id, "error", err)
		msg.Nack(false, true)
		return
	}

	ext := filepath.Ext(doc.Filename)
	imagePath := filepath.Join(uploadDir, id.String()+ext)

	text, err := processor.Process(imagePath)
	if err != nil {
		log.Error("erro ao processar OCR", "document_id", id, "error", err)
		if markErr := repository.MarkAsFailed(ctx, id, err.Error()); markErr != nil {
			log.Error("erro ao marcar como failed", "document_id", id, "error", markErr)
		}
		msg.Ack(false)
		return
	}

	documentType := domain.Classify(text)
	fields := domain.ExtractFields(text)

	if err := repository.MarkAsDone(ctx, id, text, documentType, fields); err != nil {
		log.Error("erro ao marcar como done", "document_id", id, "error", err)
		msg.Nack(false, true)
		return
	}

	log.Info("documento processado com sucesso", "document_id", id)
	msg.Ack(false)
}
