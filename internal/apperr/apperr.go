package apperr

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type AppError struct {
	Code       string
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

func FileTooLarge(err error) *AppError {
	return &AppError{
		Code:       "file_too_large",
		Message:    "arquivo maior que o limite permitido (10MB)",
		HTTPStatus: http.StatusRequestEntityTooLarge,
		Err:        err,
	}
}

func InvalidFile(err error) *AppError {
	return &AppError{
		Code:       "invalid_file",
		Message:    "arquivo inválido ou ausente no formulário",
		HTTPStatus: http.StatusBadRequest,
		Err:        err,
	}
}

func OCRFailed(err error) *AppError {
	return &AppError{
		Code:       "ocr_failed",
		Message:    "erro ao processar o documento",
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

func Respond(w http.ResponseWriter, log *slog.Logger, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		log.Warn("requisição falhou", "code", appErr.Code, "error", appErr.Err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.HTTPStatus)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   appErr.Code,
			"message": appErr.Message,
		})
		return
	}

	log.Error("erro inesperado", "error", err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "internal_error",
		"message": "erro interno no servidor",
	})
}
