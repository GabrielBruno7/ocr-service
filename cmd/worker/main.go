package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/gabrielbruno7/ocr-service/internal/document"
	"github.com/gabrielbruno7/ocr-service/internal/ocr"
	"github.com/gabrielbruno7/ocr-service/internal/platform/config"
	"github.com/gabrielbruno7/ocr-service/internal/platform/database"
	"github.com/gabrielbruno7/ocr-service/internal/platform/logger"
	"github.com/gabrielbruno7/ocr-service/internal/platform/queue"
	"github.com/google/uuid"
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

	repository := document.NewRepository(db)
	processor := ocr.NewTesseractProcessor(cfg.OCRLanguage, cfg.OCRPoolSize)

	msgs, err := q.Consume()
	if err != nil {
		log.Error("erro ao consumir fila", "error", err)
		return
	}

	log.Info("worker iniciado, aguardando mensagens...")

	for msg := range msgs {
		handleMessage(ctx, msg, repository, processor, cfg.UploadDir, log, cfg.ProcessingTimeout)
	}
}

func handleMessage(
	ctx context.Context,
	msg amqp.Delivery,
	repository *document.Repository,
	processor ocr.Processor,
	uploadDir string,
	log *slog.Logger,
	processingTimeout time.Duration,
) {
	var job document.OCRJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		log.Error("mensagem inválida, descartando", "error", err)
		msg.Nack(false, false)
		return
	}

	id, err := uuid.Parse(job.DocumentID)
	if err != nil {
		log.Error("document_id inválido", "error", err)
		msg.Nack(false, false)
		return
	}

	doc, err := repository.GetByID(ctx, id)
	if err != nil {
		log.Error("erro ao buscar documento", "document_id", id, "error", err)
		msg.Nack(false, true)
		return
	}

	if doc.Status == document.StatusDone {
		log.Info("documento já processado, ignorando", "document_id", id)
		msg.Ack(false)
		return
	}

	if doc.Status == document.StatusProcessing {
		if doc.ProcessingStartedAt != nil && time.Since(*doc.ProcessingStartedAt) < processingTimeout {
			log.Info("documento já está sendo processado por outro worker, ignorando",
				"document_id", id, "processing_started_at", doc.ProcessingStartedAt)
			msg.Ack(false)
			return
		}
		log.Warn("documento estava em processamento há muito tempo, reprocessando",
			"document_id", id, "processing_started_at", doc.ProcessingStartedAt)
	}

	log.Info("processando documento", "document_id", id, "filename", doc.Filename)

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

	if err := repository.MarkAsDone(ctx, id, text); err != nil {
		log.Error("erro ao marcar como done", "document_id", id, "error", err)
		msg.Nack(false, true)
		return
	}

	log.Info("documento processado com sucesso", "document_id", id, "text_length", len(text))
	msg.Ack(false)
}
