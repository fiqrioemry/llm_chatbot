package routes

import (
	"server/internal/config"
	"server/internal/handlers"
	"server/internal/middleware"

	"github.com/gin-gonic/gin"
)

 
func InitAuthRoutes(
	rg *gin.RouterGroup,
	cfg *config.Config,
	handlers *handlers.Handlers,
) {
	protect := middleware.Protect(cfg)
	publicAuth := rg.Group("/auth")
	protectAuth := publicAuth.Group("/auth").Use(protect)	
	{
		// Public
		publicAuth.POST("/register",                          handlers.Auth.Register)
		publicAuth.POST("/login",                             handlers.Auth.Login)
		publicAuth.POST("/verify-email",                    handlers.Auth.VerifyEmail)
		publicAuth.POST("/resend-verification-email",       handlers.Auth.ResendVerificationEmail)
		publicAuth.POST("/forgot-password",                 handlers.Auth.ForgotPassword)
		publicAuth.POST("/reset-password",                  handlers.Auth.ResetPassword)
		publicAuth.POST("/refresh",                          handlers.Auth.RefreshToken)

		// Protected
		protectAuth.POST("/logout",               handlers.Auth.Logout)
		protectAuth.POST("/sessions/revoke",      handlers.Auth.LogoutAll)
		protectAuth.DELETE("/sessions/:sessionId",handlers.Auth.DeleteSession)
		protectAuth.GET("/sessions",                    handlers.Auth.GetSessions)
		protectAuth.PATCH("/change-password",      handlers.Auth.ChangePassword)
		protectAuth.GET("/me",                     handlers.Auth.GetMe)

		publicAuth.GET("/:provider",         handlers.Auth.OAuthRedirect)
		publicAuth.GET("/:provider/callback",handlers.Auth.OAuthCallback)
	}
}