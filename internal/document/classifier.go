package document

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type DocumentType string

const (
	DocumentTypeCNH     DocumentType = "cnh"
	DocumentTypeRG      DocumentType = "rg"
	DocumentTypeUnknown DocumentType = "unknown"
)

func Classify(extractedText string) DocumentType {
	normalized := normalizeText(extractedText)

	switch {
	case containsAny(normalized,
		"CARTEIRA NACIONAL DE HABILITACAO",
		"PERMISSAO PARA DIRIGIR",
	):
		return DocumentTypeCNH

	case containsAny(normalized,
		"REGISTRO GERAL",
		"CARTEIRA DE IDENTIDADE",
	):
		return DocumentTypeRG

	default:
		return DocumentTypeUnknown
	}
}

func normalizeText(text string) string {
	text = norm.NFD.String(strings.ToUpper(text))

	var builder strings.Builder
	builder.Grow(len(text))

	for _, r := range text {
		if unicode.Is(unicode.Mn, r) {
			continue
		}

		builder.WriteRune(r)
	}

	return builder.String()
}

func containsAny(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}
