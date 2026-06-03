package routes

import (
	"server/internal/config"
	"server/internal/handlers"
	"server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitKnowledgeBaseRoutes(rg *gin.RouterGroup, cfg *config.Config, h *handlers.Handlers) {
	protect := middleware.Protect(cfg)

	kb := rg.Group("/knowledge-bases")
	kb.Use(protect)
	{
		kb.POST("", h.KnowledgeBase.Create)
		kb.GET("", h.KnowledgeBase.List)

		kbItem := kb.Group("/:id")
		{
			kbItem.GET("", h.KnowledgeBase.GetByID)
			kbItem.PATCH("", h.KnowledgeBase.Update)
			kbItem.DELETE("", h.KnowledgeBase.Delete)

			// Documents nested here so they share the :id wildcard with KB routes
			docs := kbItem.Group("/documents")
			{
				docs.POST("", h.Document.Upload)
				docs.GET("", h.Document.List)
				docs.GET("/:docId", h.Document.GetByID)
				docs.DELETE("/:docId", h.Document.Delete)
			}
		}
	}
}
