package ocr

import (
	"fmt"

	"github.com/otiai10/gosseract/v2"
)

type tesseractProcessor struct {
	languages string
}

func NewTesseractProcessor(languages string) Processor {
	return &tesseractProcessor{languages: languages}
}

func (t *tesseractProcessor) Process(imagePath string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	if err := client.SetLanguage(t.languages); err != nil {
		return "", fmt.Errorf("erro ao definir idioma: %w", err)
	}

	if err := client.SetImage(imagePath); err != nil {
		return "", fmt.Errorf("erro ao carregar imagem: %w", err)
	}

	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("erro ao extrair texto: %w", err)
	}

	return text, nil
}
