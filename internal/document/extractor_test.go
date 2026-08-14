package document

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
