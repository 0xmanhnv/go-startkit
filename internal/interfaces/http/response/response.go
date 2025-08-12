package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type Envelope struct {
	Data  any        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
	Meta  any        `json:"meta,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Data: data})
}

func BadRequest(c *gin.Context, code, msg string) {
	c.JSON(http.StatusBadRequest, Envelope{Error: &ErrorBody{Code: code, Message: msg}})
}

func Unauthorized(c *gin.Context, code, msg string) {
	c.JSON(http.StatusUnauthorized, Envelope{Error: &ErrorBody{Code: code, Message: msg}})
}

func InternalError(c *gin.Context, code, msg string) {
	c.JSON(http.StatusInternalServerError, Envelope{Error: &ErrorBody{Code: code, Message: msg}})
}

func NotFound(c *gin.Context, code, msg string) {
	c.JSON(http.StatusNotFound, Envelope{Error: &ErrorBody{Code: code, Message: msg}})
}

func Conflict(c *gin.Context, code, msg string) {
	c.JSON(http.StatusConflict, Envelope{Error: &ErrorBody{Code: code, Message: msg}})
}

// TooManyRequests sends 429 with standard envelope.
func TooManyRequests(c *gin.Context, code, msg string) {
	c.JSON(http.StatusTooManyRequests, Envelope{Error: &ErrorBody{Code: code, Message: msg}})
}

// WithDetails allows attaching details into an existing error envelope.
func WithDetails(err *ErrorBody, details any) *ErrorBody {
	if err == nil {
		return nil
	}
	err.Details = details
	return err
}

// BadRequestWithDetails sends 400 with error.details array/object.
func BadRequestWithDetails(c *gin.Context, code, msg string, details any) {
	c.JSON(http.StatusBadRequest, Envelope{Error: &ErrorBody{Code: code, Message: msg, Details: details}})
}

// PayloadTooLarge sends 413 with standard envelope.
func PayloadTooLarge(c *gin.Context, code, msg string) {
	c.JSON(http.StatusRequestEntityTooLarge, Envelope{Error: &ErrorBody{Code: code, Message: msg}})
}
