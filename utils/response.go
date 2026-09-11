package utils

import (
	"github.com/gin-gonic/gin"
)

// ResponseFormat is the standard JSON structure for API responses
type ResponseFormat struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse returns a HTTP 200/201 JSON response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, ResponseFormat{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// ErrorResponse returns an HTTP error JSON response
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ResponseFormat{
		Status:  "error",
		Message: message,
		Data:    nil,
	})
}
