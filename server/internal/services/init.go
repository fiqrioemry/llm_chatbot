package services

import (
	"server/internal/config"
	"server/internal/lib"
	"server/internal/repositories"
	"server/pkg/cache"
	"server/providers/openai"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type Service struct {
	AuthService          *AuthService
	UserService          *UserService
	FileService          *FileService
	KnowledgeBaseService *KnowledgeBaseService
	DocumentService      *DocumentService
	ChatService          *ChatService
}

func InitServices(repo *repositories.Repository, cfg *config.Config, cacheClient *cache.Client, mailer *lib.Mailer, oauth *lib.OAuthRegistry, storage *minio.Client, db *gorm.DB, ai *openai.Client) *Service {
	fileService          := NewFileService(repo.File, storage, cfg)
	userService          := NewUserService(repo.User, fileService, cacheClient)
	authService          := NewAuthService(repo.Auth, repo.User, cfg, cacheClient, mailer, oauth, db)
	knowledgeBaseService := NewKnowledgeBaseService(repo.KnowledgeBase, cfg)
	documentService      := NewDocumentService(repo.Document, repo.KnowledgeBase, ai, storage, cfg)
	chatService          := NewChatService(repo.Chat, repo.KnowledgeBase, ai, cfg)

	return &Service{
		AuthService:          authService,
		UserService:          userService,
		FileService:          fileService,
		KnowledgeBaseService: knowledgeBaseService,
		DocumentService:      documentService,
		ChatService:          chatService,
	}
}