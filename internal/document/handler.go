package document

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gabrielbruno7/ocr-service/internal/apperr"
	"github.com/gabrielbruno7/ocr-service/internal/ocr"
)

const maxUploadSize = 10 << 20

type Handler struct {
	processor  ocr.Processor
	repository *Repository
	log        *slog.Logger
}

func NewHandler(
	processor ocr.Processor,
	repository *Repository,
	log *slog.Logger,
) *Handler {
	return &Handler{
		processor:  processor,
		repository: repository,
		log:        log,
	}
}

type extractResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func (h *Handler) Extract(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

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

	tmpPath, err := saveTempFile(file, header.Filename)
	if err != nil {
		apperr.Respond(w, h.log, apperr.OCRFailed(err))
		return
	}
	defer os.Remove(tmpPath)

	text, err := h.processor.Process(tmpPath)
	if err != nil {
		apperr.Respond(w, h.log, apperr.OCRFailed(err))
		return
	}

	id, err := h.repository.Create(r.Context(), header.Filename, text)
	if err != nil {
		apperr.Respond(w, h.log, apperr.OCRFailed(err))
		return
	}

	h.log.Info("ocr concluído",
		"filename", header.Filename,
		"duration_ms", time.Since(start).Milliseconds(),
		"text_length", len(text),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(extractResponse{ID: id.String(), Text: text})
}

func saveTempFile(src io.Reader, originalName string) (string, error) {
	if err := os.MkdirAll("tmp/uploads", 0755); err != nil {
		return "", err
	}

	dstPath := filepath.Join("tmp/uploads", originalName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return dstPath, nil
}
