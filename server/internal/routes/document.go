package routes

import (
	"server/internal/config"
	"server/internal/handlers"
	"server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitDocumentRoutes(rg *gin.RouterGroup, cfg *config.Config, h *handlers.Handlers) {
	docs := rg.Group("/knowledge-bases/:kbId/documents")
	docs.Use(middleware.Protect(cfg))
	{
		docs.POST("", h.Document.Upload)
		docs.GET("", h.Document.List)
		docs.GET("/:id", h.Document.GetByID)
		docs.DELETE("/:id", h.Document.Delete)
	}
}
