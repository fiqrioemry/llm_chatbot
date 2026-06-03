package lib

import (
	"log"

	"server/internal/config"
	model "server/internal/models"

	"gorm.io/driver/mysql"
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

	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), gormCfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MySQL: %v", err)
	}

	// Auto migrate semua model
	if err := db.AutoMigrate(
		&model.User{},
		&model.UserVerification{},
		&model.ProviderAccount{},
		&model.UserActivityLog{},
		&model.Session{},
		&model.FileStorage{},
	); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}

	log.Println("✅ MySQL connected and migrated")
	return db
}