package routes

import (
	"github.com/gabrielbruno7/ocr-service/internal/document"
	"github.com/gin-gonic/gin"
)

func New(documentHandler *document.Handler) *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) { c.String(200, "ok") })

	documents := router.Group("/documents")
	{
		documents.POST("/extract", documentHandler.Extract)
		documents.GET("/:id", documentHandler.Get)
	}

	return router
}
