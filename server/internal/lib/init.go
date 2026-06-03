package lib

import (
	"server/internal/config"
	"server/pkg/cache"
	"server/pkg/validator"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)


 
 
 
func Init(cfg *config.Config) (*gorm.DB, *cache.Client, *minio.Client, *Mailer, *OAuthRegistry) {
	db     := NewDatabase(cfg)  
	redisClient := NewRedis(cfg)     
	minioClient := NewMinio(cfg)    
	mailer := NewMailer(cfg)
	oauth := NewOAuthRegistry(cfg)
	cacheClient := cache.New(redisClient)
	validator.Init()

	return  db, cacheClient, minioClient, mailer, oauth
}