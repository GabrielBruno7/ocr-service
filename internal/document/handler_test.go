package document

import (
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestShouldExtractImageTextSuccessfully(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newMemoryRepository()

	text := "texto extraído de teste"
	docID := uuid.New()
	repo.seed(&Document{
		ID:            docID,
		Status:        StatusDone,
		Filename:      "teste.png",
		ExtractedText: &text,
	})

	handler := NewHandler(repo, nil, "/tmp", slog.Default())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/documents/"+docID.String(), nil)
	c.Params = []gin.Param{{Key: "id", Value: docID.String()}}

	handler.Get(c)

	assert.Equal(t, 200, w.Code)

	var response documentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, docID.String(), response.ID)
	assert.Equal(t, StatusDone, response.Status)
	assert.Equal(t, text, *response.ExtractedText)
}

func TestSholdThrowAnErrorWhenNotFoundAnyDocumentForProvidedId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newMemoryRepository()

	handler := NewHandler(repo, nil, "/tmp", slog.Default())

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
