package ocr

import (
	"fmt"
	"image"
	"path/filepath"

	"github.com/disintegration/imaging"
)

func Preprocess(inputPath string) (string, error) {
	img, err := imaging.Open(inputPath)
	if err != nil {
		return "", fmt.Errorf("erro ao abrir imagem: %w", err)
	}

	gray := imaging.Grayscale(img)
	contrasted := imaging.AdjustContrast(gray, 30)
	resized := upscale(contrasted)

	outputPath := buildPreprocessedPath(inputPath)
	if err := imaging.Save(resized, outputPath); err != nil {
		return "", fmt.Errorf("erro ao salvar imagem pré-processada: %w", err)
	}

	return outputPath, nil
}

func buildPreprocessedPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := inputPath[:len(inputPath)-len(ext)]
	return base + "_preprocessed" + ext
}

func upscale(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()

	const minWidth = 1500

	if width >= minWidth {
		return img
	}

	return imaging.Resize(img, minWidth, 0, imaging.Lanczos)
}
