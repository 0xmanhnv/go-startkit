package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func registerHealthRoutes(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "path": c.Request.URL.Path})
	})
}
