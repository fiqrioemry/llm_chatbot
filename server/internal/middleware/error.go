package middleware

import (
	"log"
	"net/http"

	"server/internal/config/constant"
	"server/pkg/response"

	"github.com/gin-gonic/gin"
)


func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[panic] %v", err)
				response.InternalError(c,
					constant.ErrInternalServerError,
					constant.CodeInternalServerError,
				)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// NotFound handler untuk route yang tidak terdaftar.
func NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": c.Request.Method + " " + c.Request.URL.Path,
			"code":    constant.CodeRouteNotFound,
		})
	}
}