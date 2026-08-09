package main

import (
	"fmt"
	"net/http"

	"github.com/gabrielbruno7/ocr-service/internal/document"
	"github.com/gabrielbruno7/ocr-service/internal/ocr"
)

func main() {
	processor := ocr.NewTesseractProcessor("por")
	docHandler := document.NewHandler(processor)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	http.HandleFunc("/documents/extract", docHandler.Extract)

	fmt.Println("servidor rodando em :8080")
	http.ListenAndServe(":8080", nil)
}
