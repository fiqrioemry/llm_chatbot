package handlers

import (
	"server/internal/config"
	"server/internal/services"
)

type Handlers struct {
	Auth          *AuthHandler
	User          *UserHandler
	KnowledgeBase *KnowledgeBaseHandler
	Document      *DocumentHandler
	Chat          *ChatHandler
}

func InitHandlers(s *services.Service, cfg *config.Config) *Handlers {
	return &Handlers{
		Auth:          NewAuthHandler(s.AuthService, cfg),
		User:          NewUserHandler(s.UserService, cfg),
		KnowledgeBase: NewKnowledgeBaseHandler(s.KnowledgeBaseService, cfg),
		Document:      NewDocumentHandler(s.DocumentService, cfg),
		Chat:          NewChatHandler(s.ChatService, cfg),
	}
}