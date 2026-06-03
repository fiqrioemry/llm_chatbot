package routes

import (
	"time"

	"server/internal/config"
	"server/internal/handlers"
	"server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoutes(
	engine *gin.Engine,
	cfg *config.Config,
	handlers *handlers.Handlers,
) {
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": 200, "message": "Hello from Go + Gin API!", "timestamp": time.Now().UTC()})
	})
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": 200, "message": "OK", "timestamp": time.Now().UTC()})
	})
	engine.NoRoute(middleware.NotFound())

	groupV1 := engine.Group("/api/v1")
	{
		InitAuthRoutes(groupV1, cfg, handlers)
		InitUserRoutes(groupV1, cfg, handlers)
	}
}