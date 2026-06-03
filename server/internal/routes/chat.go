package routes

import (
	"server/internal/config"
	"server/internal/handlers"
	"server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitChatRoutes(rg *gin.RouterGroup, cfg *config.Config, h *handlers.Handlers) {
	convs := rg.Group("/conversations")
	convs.Use(middleware.Protect(cfg))
	{
		convs.POST("", h.Chat.CreateConversation)
		convs.GET("", h.Chat.ListConversations)
		convs.DELETE("/:id", h.Chat.DeleteConversation)
		convs.POST("/:id/messages", h.Chat.SendMessage)
		convs.GET("/:id/messages", h.Chat.ListMessages)
	}
}
