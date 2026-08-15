package apperr

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type ErrorCode int

const (
	CodeFileTooLarge ErrorCode = iota + 4000
	CodeInvalidFile
	CodeOCRFailed
	CodeNotFound
)

const CodeInternal ErrorCode = 5000

type errorDefinition struct {
	Message    string
	HTTPStatus int
}

var registry = map[ErrorCode]errorDefinition{
	CodeFileTooLarge: {"arquivo maior que o limite permitido (10MB)", http.StatusRequestEntityTooLarge},
	CodeInvalidFile:  {"arquivo inválido ou ausente no formulário", http.StatusBadRequest},
	CodeOCRFailed:    {"erro ao processar o documento", http.StatusInternalServerError},
	CodeNotFound:     {"documento não encontrado", http.StatusNotFound},
	CodeInternal:     {"erro interno no servidor", http.StatusInternalServerError},
}

type AppError struct {
	Code       ErrorCode
	Message    string
	HTTPStatus int
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code ErrorCode, err error) *AppError {
	def, ok := registry[code]
	if !ok {
		def = registry[CodeInternal]
		code = CodeInternal
	}

	return &AppError{
		Code:       code,
		Message:    def.Message,
		HTTPStatus: def.HTTPStatus,
		Err:        err,
	}
}

func Respond(w http.ResponseWriter, log *slog.Logger, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		logAppError(log, appErr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.HTTPStatus)
		json.NewEncoder(w).Encode(map[string]any{
			"error":   appErr.Code,
			"message": appErr.Message,
		})
		return
	}

	log.Error("erro inesperado", "error", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]any{
		"error":   CodeInternal,
		"message": "erro interno no servidor",
	})
}

func logAppError(log *slog.Logger, appErr *AppError) {
	if appErr.HTTPStatus >= 500 {
		log.Error("requisição falhou", "code", appErr.Code, "error", appErr.Err)
	} else {
		log.Warn("requisição falhou", "code", appErr.Code, "error", appErr.Err)
	}
}
