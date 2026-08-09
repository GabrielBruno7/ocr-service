package ocr

import (
	"fmt"

	"github.com/otiai10/gosseract/v2"
)

type tesseractProcessor struct {
	pool chan *gosseract.Client
}

func NewTesseractProcessor(languages string, poolSize int) Processor {
	pool := make(chan *gosseract.Client, poolSize)

	for i := 0; i < poolSize; i++ {
		client := gosseract.NewClient()
		client.SetLanguage(languages)
		pool <- client
	}

	return &tesseractProcessor{pool: pool}
}

func (t *tesseractProcessor) Process(imagePath string) (string, error) {
	client := <-t.pool

	defer func() {
		t.pool <- client
	}()

	if err := client.SetImage(imagePath); err != nil {
		return "", fmt.Errorf("erro ao carregar imagem: %w", err)
	}

	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("erro ao extrair texto: %w", err)
	}

	return text, nil
}
