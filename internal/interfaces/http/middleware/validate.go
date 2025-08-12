package middleware

import (
	"net/http"

	resp "gostartkit/internal/interfaces/http/response"
	"gostartkit/internal/interfaces/http/validation"

	"github.com/gin-gonic/gin"
)

// ValidateJSON binds and validates request JSON into a typed payload before reaching the handler.
// The validated payload is stored in context with the provided ctxKey.
func ValidateJSON[T any](ctxKey string, maxBodyBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBodyBytes > 0 && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
		}
		var payload T
		if err := c.ShouldBindJSON(&payload); err != nil {
			if validation.IsBodyTooLarge(err) {
				resp.PayloadTooLarge(c, resp.CodePayloadTooLarge, resp.MsgPayloadTooLarge)
				return
			}
			// Prefer returning detailed list of validation errors when available
			if verrs := validation.MapValidationErrors(err); len(verrs) > 0 {
				// Localize messages using Accept-Language
				localized := validation.MapValidationErrorsWithLocale(GetLocale(c), err)
				if len(localized) == 0 {
					localized = verrs
				}
				resp.BadRequestWithDetails(c, resp.CodeInvalidRequest, resp.MsgInvalidJSON, localized)
				return
			}
			// Fallback single-error mapping with locale
			code, msg := validation.MapBindJSONErrorWithLocale(GetLocale(c), err)
			resp.BadRequest(c, code, msg)
			return
		}
		c.Set(ctxKey, payload)
		c.Next()
	}
}

// parseLocale extracts the primary language code from Accept-Language header.
// locale parsing now delegated to pkg/i18n
