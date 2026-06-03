package services

import (
	"server/internal/config"
	"server/internal/lib"
	"server/internal/repositories"
	"server/pkg/cache"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type Service struct {
	// Add any additional services here
	AuthService *AuthService
	UserService *UserService
	FileService *FileService
}

func InitServices(repo *repositories.Repository, cfg *config.Config, cacheClient *cache.Client, mailer *lib.Mailer, oauth *lib.OAuthRegistry, storage *minio.Client, db *gorm.DB) *Service {
	fileService := NewFileService(repo.File, storage, cfg)
	userService := NewUserService(repo.User, fileService, cacheClient)
	authService := NewAuthService(repo.Auth, repo.User, cfg, cacheClient, mailer, oauth, db)

	return &Service{
		AuthService: authService,
		UserService: userService,
		FileService: fileService,
	}
}