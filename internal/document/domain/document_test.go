package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldExtractCPFFromText(t *testing.T) {
	text := "NOME: JOAO DA SILVA\nCPF: 123.456.789-00\nCATEGORIA: B"

	fields := ExtractFields(text)

	assert.Equal(t, "123.456.789-00", fields.CPF)
}

func TestShouldExtractNameFromText(t *testing.T) {
	text := "NOME: JOAO DA SILVA\nCPF: 123.456.789-00"

	fields := ExtractFields(text)

	assert.Equal(t, "JOAO DA SILVA", fields.Name)
}

func TestShouldExtractNameWithSobrenomeLabel(t *testing.T) {
	text := "NOME E SOBRENOME\nMARIA TESTE SOUZA\nCPF: 000.000.000-00"

	fields := ExtractFields(text)

	assert.Equal(t, "MARIA TESTE SOUZA", fields.Name)
}

func TestShouldReturnEmptyFieldsWhenNothingMatches(t *testing.T) {
	text := "texto qualquer sem nenhum campo reconhecível"

	fields := ExtractFields(text)

	assert.Empty(t, fields.CPF)
	assert.Empty(t, fields.Name)
}

func TestShouldExtractBothFieldsTogether(t *testing.T) {
	text := "CARTEIRA NACIONAL DE HABILITAÇÃO\nNOME: JOAO DA SILVA TESTE\nCPF: 000.000.000-00\nCATEGORIA: B"

	fields := ExtractFields(text)

	assert.Equal(t, "JOAO DA SILVA TESTE", fields.Name)
	assert.Equal(t, "000.000.000-00", fields.CPF)
}

func TestShouldClassifyDocumentWithCNHType(t *testing.T) {
	text := "CARTEIRA NACIONAL DE HABILITAÇÃO\nNOME: JOAO DA SILVA\nCATEGORIA: B"

	result := Classify(text)

	assert.Equal(t, DocumentTypeCNH, result)
}

func TestShouldClassifyDocumentWithCNHTypeWithoutAccentMark(t *testing.T) {
	text := "PERMISSAO PARA DIRIGIR\nNOME: JOAO DA SILVA"

	result := Classify(text)

	assert.Equal(t, DocumentTypeCNH, result)
}

func TestShouldClassifyDocumentWithRGType(t *testing.T) {
	text := "REGISTRO GERAL\nNOME: MARIA SOUZA\nCPF: 123.456.789-00"

	result := Classify(text)

	assert.Equal(t, DocumentTypeRG, result)
}

func TestShouldClassifyDocumentWithUnknowType(t *testing.T) {
	text := "isso não é um documento reconhecido, só um texto qualquer"

	result := Classify(text)

	assert.Equal(t, DocumentTypeUnknown, result)
}

func TestShouldClassifyDocumentWithUnknowTypeWhenExtractAnEmptyText(t *testing.T) {
	result := Classify("")

	assert.Equal(t, DocumentTypeUnknown, result)
}

func TestShouldClassifyDocumentWithCNHTypeEvenWithOCRTypo(t *testing.T) {
	text := "CARTELRA NACIONAL DE HABILITAÇÃO\nDRIVER LICENSE"

	result := Classify(text)

	assert.Equal(t, DocumentTypeCNH, result)
}

func TestShouldClassifyDocumentWithCNHTypeEvenWithLineBreakInsideWord(t *testing.T) {
	text := "CARTEIRA NA\nCIONAL DE HABILITAÇÃO"

	result := Classify(text)

	assert.Equal(t, DocumentTypeCNH, result)
}

func TestShouldNotClassifyRandomTextAsCNHEvenWithFuzzyMatching(t *testing.T) {
	text := "este é um texto qualquer sem relação com documentos de identidade"

	result := Classify(text)

	assert.Equal(t, DocumentTypeUnknown, result)
}
