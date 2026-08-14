package document

import (
	"regexp"
	"strings"
)

type ExtractedFields struct {
	Name string `json:"name,omitempty"`
	CPF  string `json:"cpf,omitempty"`
}

var (
	cpfPattern = regexp.MustCompile(`\d{3}\.\d{3}\.\d{3}-\d{2}`)

	namePattern = regexp.MustCompile(`(?i)NOME(?:\s+E\s+SOBRENOME)?[:\s]+([A-ZÀ-Ú\s]+?)(?:\n|$)`)
)

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
