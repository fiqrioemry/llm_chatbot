package middleware

import (
	"log"
	"strings"

	"server/internal/config"
	"server/internal/config/constant"
	"server/pkg/jwt"
	"server/pkg/response"

	"github.com/gin-gonic/gin"
)

 
type AuthUser struct {
	UserID   string
	Username string
	Role     string
}

const ContextKeyUser = "auth_user"

 
func Protect(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[Auth Middleware] %s %s", c.Request.Method, c.Request.URL.Path)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Unauthorized(c, constant.ErrUnauthorized, constant.CodeUnauthorized)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwt.VerifyAccessToken(cfg.Auth.AccessTokenSecret, tokenStr)
		if err != nil {
			response.Unauthorized(c, constant.ErrInvalidToken, constant.CodeInvalidToken)
			c.Abort()
			return
		}

		c.Set(ContextKeyUser, &AuthUser{
			UserID:   claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		})

		c.Next()
	}
}

// RequireRole memastikan user yang sudah ter-autentikasi memiliki role tertentu.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetAuthUser(c)
		if user == nil {
			response.Unauthorized(c, constant.ErrUnauthorized, constant.CodeUnauthorized)
			c.Abort()
			return
		}

		for _, r := range roles {
			if user.Role == r {
				c.Next()
				return
			}
		}

		response.Forbidden(c, constant.ErrForbidden, constant.CodeForbidden)
		c.Abort()
	}
}

// GetAuthUser mengambil AuthUser dari context — dipakai di handler.
func GetAuthUser(c *gin.Context) *AuthUser {
	v, exists := c.Get(ContextKeyUser)
	if !exists {
		return nil
	}
	user, ok := v.(*AuthUser)
	if !ok {
		return nil
	}
	return user
}