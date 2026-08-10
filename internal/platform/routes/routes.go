package routes

import (
	"log/slog"
	"net/http"

	"github.com/gabrielbruno7/ocr-service/internal/document"
)

func New(docHandler *document.Handler, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/documents/extract", docHandler.Extract)

	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
