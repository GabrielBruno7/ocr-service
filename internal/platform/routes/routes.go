package routes

import "github.com/gin-gonic/gin"

type RouteRegister interface {
	RegisterRoutes(router *gin.Engine)
}

func New(handlers ...RouteRegister) *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) { c.String(200, "ok") })

	for _, r := range handlers {
		r.RegisterRoutes(router)
	}

	return router
}
