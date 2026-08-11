package apperr

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileTooLarge(t *testing.T) {
	originalErr := errors.New("body too large")

	err := New(CodeFileTooLarge, originalErr)

	assert.Equal(t, CodeFileTooLarge, err.Code)
	assert.Equal(t, http.StatusRequestEntityTooLarge, err.HTTPStatus)
	assert.Equal(t, "arquivo maior que o limite permitido (10MB)", err.Message)
	assert.Equal(t, originalErr, err.Err)
}

func TestInvalidFile(t *testing.T) {
	err := New(CodeInvalidFile, nil)

	assert.Equal(t, CodeInvalidFile, err.Code)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
	assert.Equal(t, "arquivo inválido ou ausente no formulário", err.Message)
}

func TestOCRFailed(t *testing.T) {
	originalErr := errors.New("tesseract crashed")

	err := New(CodeOCRFailed, originalErr)

	assert.Equal(t, CodeOCRFailed, err.Code)
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus)
}

func TestNotFound(t *testing.T) {
	err := New(CodeNotFound, nil)

	assert.Equal(t, CodeNotFound, err.Code)
	assert.Equal(t, http.StatusNotFound, err.HTTPStatus)
}

func TestUnknownCodeFallsBackToInternal(t *testing.T) {
	err := New(ErrorCode(9999), nil)

	assert.Equal(t, CodeInternal, err.Code)
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus)
}
