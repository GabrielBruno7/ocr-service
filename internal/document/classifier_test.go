package document

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
