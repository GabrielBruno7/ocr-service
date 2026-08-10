package document

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gabrielbruno7/ocr-service/internal/apperr"
	"github.com/gabrielbruno7/ocr-service/internal/platform/queue"
)

const maxUploadSize = 10 << 20

type Handler struct {
	repository *Repository
	queue      *queue.Queue
	uploadDir  string
	log        *slog.Logger
}

func NewHandler(repository *Repository, q *queue.Queue, uploadDir string, log *slog.Logger) *Handler {
	return &Handler{repository: repository, queue: q, uploadDir: uploadDir, log: log}
}

type extractResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (h *Handler) Extract(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		if err.Error() == "http: request body too large" {
			apperr.Respond(w, h.log, apperr.FileTooLarge(err))
			return
		}
		apperr.Respond(w, h.log, apperr.InvalidFile(err))
		return
	}
	defer file.Close()

	h.log.Info("upload recebido", "filename", header.Filename, "size_bytes", header.Size)

	id, err := h.repository.CreatePending(r.Context(), header.Filename)
	if err != nil {
		apperr.Respond(w, h.log, apperr.OCRFailed(err))
		return
	}

	if err := h.saveUploadedFile(file, id.String(), header.Filename); err != nil {
		apperr.Respond(w, h.log, apperr.OCRFailed(err))
		return
	}

	if err := PublishOCRJob(r.Context(), h.queue, id.String()); err != nil {
		apperr.Respond(w, h.log, apperr.OCRFailed(err))
		return
	}

	h.log.Info("documento enfileirado", "document_id", id, "filename", header.Filename)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(extractResponse{ID: id.String(), Status: "pending"})
}

func (h *Handler) saveUploadedFile(src io.Reader, documentID, originalName string) error {
	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de uploads: %w", err)
	}

	ext := filepath.Ext(originalName)
	dstPath := filepath.Join(h.uploadDir, documentID+ext)

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("erro ao copiar arquivo: %w", err)
	}

	return nil
}
