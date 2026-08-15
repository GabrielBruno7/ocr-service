package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ocr-service/ocr-service/internal/document/domain"
	"github.com/ocr-service/ocr-service/internal/document/usecase"
)

func TestShouldExtractImageTextSuccessfully(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := domain.NewMemoryRepository()

	text := "texto extraído de teste"
	docID := uuid.New()
	repo.Seed(&domain.Document{
		ID:            docID,
		Status:        domain.StatusDone,
		Filename:      "teste.png",
		ExtractedText: &text,
	})

	ctx := context.Background()
	getDocumentUsecase := usecase.NewGetDocumentUsecase(ctx, repo)
	handler := NewDocumentHandler(slog.Default(), getDocumentUsecase, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/documents/"+docID.String(), nil)
	c.Params = []gin.Param{{Key: "id", Value: docID.String()}}

	handler.Get(c)

	assert.Equal(t, 200, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, docID.String(), response["id"])
	assert.Equal(t, string(domain.StatusDone), response["status"])
	assert.Equal(t, text, response["extracted_text"])
}

func TestSholdThrowAnErrorWhenNotFoundAnyDocumentForProvidedId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := domain.NewMemoryRepository()

	ctx := context.Background()
	getDocumentUsecase := usecase.NewGetDocumentUsecase(ctx, repo)
	handler := NewDocumentHandler(slog.Default(), getDocumentUsecase, nil)

	unknownID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/documents/"+unknownID.String(), nil)
	c.Params = []gin.Param{{Key: "id", Value: unknownID.String()}}

	handler.Get(c)

	assert.Equal(t, 404, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "documento não encontrado", response["message"])
}
