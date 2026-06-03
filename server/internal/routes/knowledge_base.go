package routes

import (
	"server/internal/config"
	"server/internal/handlers"
	"server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitKnowledgeBaseRoutes(rg *gin.RouterGroup, cfg *config.Config, h *handlers.Handlers) {
	kb := rg.Group("/knowledge-bases")
	kb.Use(middleware.Protect(cfg))
	{
		kb.POST("", h.KnowledgeBase.Create)
		kb.GET("", h.KnowledgeBase.List)
		kb.GET("/:id", h.KnowledgeBase.GetByID)
		kb.PATCH("/:id", h.KnowledgeBase.Update)
		kb.DELETE("/:id", h.KnowledgeBase.Delete)
	}
}
