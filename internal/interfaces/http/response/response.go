package response

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type ErrorBody struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

type Envelope struct {
    Data  interface{} `json:"data,omitempty"`
    Error *ErrorBody  `json:"error,omitempty"`
    Meta  interface{} `json:"meta,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Envelope{Data: data})
}

func Created(c *gin.Context, data interface{}) {
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

