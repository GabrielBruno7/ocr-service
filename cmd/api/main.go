package main

import (
	"net/http"

	"github.com/gabrielbruno7/ocr-service/internal/document"
	"github.com/gabrielbruno7/ocr-service/internal/ocr"
	"github.com/gabrielbruno7/ocr-service/internal/platform/config"
	"github.com/gabrielbruno7/ocr-service/internal/platform/logger"
)

func main() {
	log := logger.New()
	cfg := config.Load()

	processor := ocr.NewTesseractProcessor(cfg.OCRLanguage, cfg.OCRPoolSize)
	docHandler := document.NewHandler(processor, log)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/documents/extract", docHandler.Extract)

	log.Info("servidor iniciado", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Error("erro ao iniciar servidor", "error", err)
	}
}
