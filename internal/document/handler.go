package document

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gabrielbruno7/ocr-service/internal/apperr"
	"github.com/gabrielbruno7/ocr-service/internal/platform/queue"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxUploadSize = 10 << 20

type Handler struct {
	repository Repository
	queue      queue.Publisher
	uploadDir  string
	log        *slog.Logger
}

func NewHandler(repository Repository, q queue.Publisher, uploadDir string, log *slog.Logger) *Handler {
	return &Handler{repository: repository, queue: q, uploadDir: uploadDir, log: log}
}

type extractResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (h *Handler) Extract(c *gin.Context) {
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

	id, err := h.repository.CreatePending(c.Request.Context(), header.Filename)
	if err != nil {
		apperr.Respond(c.Writer, h.log, apperr.New(apperr.CodeOCRFailed, err))
		return
	}

	if err := h.saveUploadedFile(file, id.String(), header.Filename); err != nil {
		apperr.Respond(c.Writer, h.log, apperr.New(apperr.CodeOCRFailed, err))
		return
	}

	if err := PublishOCRJob(c.Request.Context(), h.queue, id.String()); err != nil {
		apperr.Respond(c.Writer, h.log, apperr.New(apperr.CodeOCRFailed, err))
		return
	}

	h.log.Info("documento enfileirado", "document_id", id, "filename", header.Filename)

	c.JSON(http.StatusAccepted, extractResponse{ID: id.String(), Status: "pending"})
}

type documentResponse struct {
	ID              string           `json:"id"`
	Status          string           `json:"status"`
	Filename        string           `json:"filename"`
	DocumentType    *string          `json:"document_type,omitempty"`
	ExtractedText   *string          `json:"extracted_text,omitempty"`
	ExtractedFields *ExtractedFields `json:"extracted_fields,omitempty"`
	ErrorMessage    *string          `json:"error_message,omitempty"`
}

func (h *Handler) Get(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		apperr.Respond(c.Writer, h.log, apperr.New(apperr.CodeInvalidFile, err))
		return
	}

	doc, err := h.repository.GetByID(c.Request.Context(), id)
	if err != nil {
		apperr.Respond(c.Writer, h.log, err)
		return
	}

	c.JSON(http.StatusOK, documentResponse{
		ID:              doc.ID.String(),
		Status:          doc.Status,
		Filename:        doc.Filename,
		DocumentType:    doc.DocumentType,
		ExtractedText:   doc.ExtractedText,
		ExtractedFields: doc.ExtractedFields,
		ErrorMessage:    doc.ErrorMessage,
	})
}

func (h *Handler) saveUploadedFile(src io.Reader, documentID, originalName string) error {
	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de uploads: %w", err)
	}

	ext := filepath.Ext(originalName)
	dstPath := filepath.Join(h.uploadDir, documentID+ext)

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("erro ao copiar arquivo: %w", err)
	}

	return nil
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	documents := router.Group("/documents")
	{
		documents.POST("/extract", h.Extract)
		documents.GET("/:id", h.Get)
	}
}
