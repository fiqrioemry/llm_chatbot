package routes

import (
	"server/internal/config"
	"server/internal/handlers"

	"github.com/gin-gonic/gin"
)

// InitDocumentRoutes is a no-op — document routes are registered inside
// InitKnowledgeBaseRoutes to share the :id wildcard with KB routes.
func InitDocumentRoutes(_ *gin.RouterGroup, _ *config.Config, _ *handlers.Handlers) {}
