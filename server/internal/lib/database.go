package lib

import (
	"log"

	"server/internal/config"
	model "server/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(cfg *config.Config) *gorm.DB {
	gormCfg := &gorm.Config{}

	if cfg.App.Mode == "development" {
		gormCfg.Logger = logger.Default.LogMode(logger.Info)
	} else {
		gormCfg.Logger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), gormCfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}

	// Enable pgvector extension before AutoMigrate so vector columns are recognised
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		log.Fatalf("❌ Failed to enable pgvector extension: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.UserVerification{},
		&model.ProviderAccount{},
		&model.UserActivityLog{},
		&model.Session{},
		&model.FileStorage{},
		&model.KnowledgeBase{},
		&model.Document{},
		&model.DocumentChunk{},
		&model.Conversation{},
		&model.Message{},
		&model.MessageSource{},
	); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}

	log.Println("✅ PostgreSQL connected and migrated")
	return db
}
