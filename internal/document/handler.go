package document

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gabrielbruno7/ocr-service/internal/ocr"
)

type Handler struct {
	processor ocr.Processor
	log       *slog.Logger
}

func NewHandler(processor ocr.Processor, log *slog.Logger) *Handler {
	return &Handler{processor: processor, log: log}
}

type extractResponse struct {
	Text string `json:"text"`
}

func (h *Handler) Extract(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	file, header, err := r.FormFile("file")
	if err != nil {
		h.log.Warn("arquivo não encontrado no formulário", "error", err)
		http.Error(w, "arquivo 'file' não encontrado no formulário", http.StatusBadRequest)
		return
	}
	defer file.Close()

	h.log.Info("upload recebido", "filename", header.Filename, "size_bytes", header.Size)

	tmpPath, err := saveTempFile(file, header.Filename)
	if err != nil {
		http.Error(w, "erro ao salvar arquivo temporário", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpPath)

	text, err := h.processor.Process(tmpPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("erro ao processar OCR: %v", err), http.StatusInternalServerError)
		return
	}

	h.log.Info("ocr concluído",
		"filename", header.Filename,
		"duration_ms", time.Since(start).Milliseconds(),
		"text_length", len(text),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(extractResponse{Text: text})
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
