package response

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
)

 

type AppError struct {
	Status  int
	Message string
	Code    string
}

func (e *AppError) Error() string { return e.Message }


func NewError(status int, message, code string) *AppError {
	return &AppError{Status: status, Message: message, Code: code}
}

func BadRequestErr(message, code string) *AppError {
	return NewError(400, message, code)
}

func UnauthorizedErr(message, code string) *AppError {
	return NewError(401, message, code)
}

func ForbiddenErr(message, code string) *AppError {
	return NewError(403, message, code)
}

func NotFoundErr(message, code string) *AppError {
	return NewError(404, message, code)
}

func InternalErr(message, code string) *AppError {
	return NewError(500, message, code)
}

// ─── HandleError ──────────────────────────────────────────────────────────────
// Global error dispatcher — dipanggil di semua handler, tidak perlu
// di-copy per module. Menggantikan handleError(c, err) lokal di tiap handler.
//
// Penggunaan di handler:
//   result, err := h.svc.Login(c, req)
//   if err != nil {
//       response.HandleError(c, err)
//       return
//   }

func HandleError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		// Domain error — status dan pesan sudah ditentukan di service
		Error(c, appErr.Status, appErr.Message, appErr.Code)
		return
	}

	// Unexpected error — log untuk investigasi, jangan expose detail ke client
	log.Printf("[ERROR] unhandled error on %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	InternalError(c, "Internal server error", "INTERNAL_SERVER_ERROR")
}

// ─── Envelope structs ─────────────────────────────────────────────────────────

type successBody struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type errorBody struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Errors  interface{} `json:"errors,omitempty"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

// ─── Success responses ────────────────────────────────────────────────────────

func OK(c *gin.Context, message string, data ...interface{}) {
	body := successBody{Success: true, Message: message}
	if len(data) > 0 {
		body.Data = data[0]
	}
	if len(data) > 1 {
		body.Meta = data[1]
	}
	c.JSON(200, body)
}

func Created(c *gin.Context, message string, data ...interface{}) {
	body := successBody{Success: true, Message: message}
	if len(data) > 0 {
		body.Data = data[0]
	}
	c.JSON(201, body)
}

// ─── Error responses ──────────────────────────────────────────────────────────

func Error(c *gin.Context, status int, message, code string, errors ...interface{}) {
	body := errorBody{
		Success: false,
		Message: message,
		Code:    code,
	}
	if len(errors) > 0 {
		body.Errors = errors[0]
	}
	c.JSON(status, body)
}

func Unauthorized(c *gin.Context, message, code string) {
	Error(c, 401, message, code)
}

func Forbidden(c *gin.Context, message, code string) {
	Error(c, 403, message, code)
}

func NotFound(c *gin.Context, message, code string) {
	Error(c, 404, message, code)
}

func BadRequest(c *gin.Context, message, code string, errors ...interface{}) {
	Error(c, 400, message, code, errors...)
}

func TooManyRequests(c *gin.Context, message, code string) {
	Error(c, 429, message, code)
}

func InternalError(c *gin.Context, message, code string) {
	Error(c, 500, message, code)
}