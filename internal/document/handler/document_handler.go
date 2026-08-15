package handler

import (
	"log/slog"
	"net/http"

	"github.com/ocr-service/ocr-service/internal/apperr"
	"github.com/ocr-service/ocr-service/internal/document/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxUploadSize = 10 << 20

type DocumentHandler struct {
	log                    *slog.Logger
	getDocumentUsecase     *usecase.GetDocumentUsecase
	extractDocumentUsecase *usecase.ExtractDocumentUsecase
}

func NewDocumentHandler(
	log *slog.Logger,
	getDocumentUsecase *usecase.GetDocumentUsecase,
	extractDocumentUsecase *usecase.ExtractDocumentUsecase,
) *DocumentHandler {
	return &DocumentHandler{
		log:                    log,
		getDocumentUsecase:     getDocumentUsecase,
		extractDocumentUsecase: extractDocumentUsecase,
	}
}

func (h *DocumentHandler) Extract(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if err.Error() == "http: request body too large" {
			apperr.Respond(c.Writer, h.log, apperr.New(apperr.CodeFileTooLarge, err))
			return
		}
		apperr.Respond(c.Writer, h.log, apperr.New(apperr.CodeInvalidFile, err))
		return
	}
	defer file.Close()

	h.log.Info("upload recebido", "filename", header.Filename, "size_bytes", header.Size)

	input := usecase.ExtractDocumentInput{
		File:     file,
		FileName: header.Filename,
		FileSize: header.Size,
	}

	id, appErr := h.extractDocumentUsecase.Execute(input)
	if appErr != nil {
		apperr.Respond(c.Writer, h.log, appErr)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"id":     id,
		"status": "pending",
	})
}

func (h *DocumentHandler) Get(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		apperr.Respond(c.Writer, h.log, apperr.New(apperr.CodeInvalidFile, err))
		return
	}

	doc, appErr := h.getDocumentUsecase.Execute(id)
	if appErr != nil {
		apperr.Respond(c.Writer, h.log, appErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               doc.ID.String(),
		"status":           doc.Status,
		"file_name":        doc.Filename,
		"document_type":    doc.DocumentType,
		"extracted_text":   doc.ExtractedText,
		"extracted_fields": doc.ExtractedFields,
		"error_message":    doc.ErrorMessage,
	})
}

func (h *DocumentHandler) RegisterRoutes(router *gin.Engine) {
	documents := router.Group("/documents")
	{
		documents.POST("/extract", h.Extract)
		documents.GET("/:id", h.Get)
	}
}
