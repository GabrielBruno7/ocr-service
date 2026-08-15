package domain

import (
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/agnivade/levenshtein"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusDone       = "done"
	StatusFailed     = "failed"
)

type DocumentType string

var (
	cpfPattern = regexp.MustCompile(`\d{3}\.\d{3}\.\d{3}-\d{2}`)

	namePattern = regexp.MustCompile(`(?i)NOME(?:\s+E\s+SOBRENOME)?[:\s]+([A-ZÀ-Ú\s]+?)(?:\n|$)`)
)

const (
	DocumentTypeCNH     DocumentType = "cnh"
	DocumentTypeRG      DocumentType = "rg"
	DocumentTypeUnknown DocumentType = "unknown"
)

type ExtractedFields struct {
	Name string `json:"name,omitempty"`
	CPF  string `json:"cpf,omitempty"`
}

type Document struct {
	ID                  uuid.UUID
	Status              string
	Filename            string
	DocumentType        *string
	ExtractedText       *string
	ExtractedFields     *ExtractedFields
	ErrorMessage        *string
	ProcessingStartedAt *time.Time
}

const maxEditDistanceRatio = 0.2

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

	return strings.Join(strings.Fields(builder.String()), " ")
}

func containsAny(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if fuzzyContains(text, keyword) {
			return true
		}
	}
	return false
}

func fuzzyContains(text, keyword string) bool {
	keywordLen := len(keyword)
	maxDistance := int(float64(keywordLen) * maxEditDistanceRatio)

	if len(text) < keywordLen {
		return levenshtein.ComputeDistance(text, keyword) <= maxDistance
	}

	for i := 0; i+keywordLen <= len(text); i++ {
		window := text[i : i+keywordLen]
		if levenshtein.ComputeDistance(window, keyword) <= maxDistance {
			return true
		}
	}

	return false
}

func ExtractFields(extractedText string) ExtractedFields {
	fields := ExtractedFields{}

	if match := cpfPattern.FindString(extractedText); match != "" {
		fields.CPF = match
	}

	if match := namePattern.FindStringSubmatch(extractedText); len(match) > 1 {
		fields.Name = strings.TrimSpace(match[1])
	}

	return fields
}
