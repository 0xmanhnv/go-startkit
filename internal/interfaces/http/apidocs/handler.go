package apidocs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Mount registers /swagger and /openapi.json under the given router group.
// Only mount this in dev environments.
func Mount(r *gin.Engine) {
	r.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", OpenAPISpec)
	})
	// Minimal static viewer: redirect /swagger to external ReDoc CDN via HTML wrapper
	r.GET("/swagger", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `<!DOCTYPE html><html><head><title>API Docs</title></head><body>
<redoc spec-url='/openapi.json'></redoc>
<script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body></html>`)
	})
}
