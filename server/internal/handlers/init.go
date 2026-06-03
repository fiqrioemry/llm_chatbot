package handlers

import (
	"server/internal/config"
	"server/internal/services"
)

type Handlers struct {
	// Add any dependencies or services needed by the handlers here
	Auth *AuthHandler
	User *UserHandler
}

func InitHandlers(s *services.Service, cfg *config.Config) *Handlers {
	return &Handlers{
		Auth : NewAuthHandler(s.AuthService, cfg),
		User : NewUserHandler(s.UserService, cfg),
		// Initialize any dependencies or services here
	}
}