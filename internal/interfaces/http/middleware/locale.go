package middleware

import (
	"gostartkit/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// LocaleMiddleware captures Accept-Language and stores locale in Gin context (decouples pkg/i18n from Gin).
func LocaleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		loc := i18n.ParseAcceptLanguage(c.GetHeader("Accept-Language"))
		c.Set("locale", loc)
		c.Next()
	}
}

// GetLocale returns request locale (from context), or default.
func GetLocale(c *gin.Context) string {
	if v, ok := c.Get("locale"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return i18n.DefaultLocale()
}
