package routes

import (
	"server/internal/config"
	"server/internal/handlers"
	"server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitUserRoutes(
	rg *gin.RouterGroup,
	cfg *config.Config,
	handlers *handlers.Handlers,
) {
	protect := middleware.Protect(cfg)

	u := rg.Group("/user")
	{
		u.PUT("/profile", protect, handlers.User.UpdateProfile)
		u.PATCH("/profile/avatar", protect, handlers.User.UpdateAvatar)
	}
}