package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileStorage struct {
	ID        string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TargetID  string          `gorm:"type:uuid;not null;index"`
	Module    string          `gorm:"type:varchar(50);not null;index"`
	URL       string          `gorm:"type:varchar(500);not null"`
	Path      string          `gorm:"type:varchar(500);not null"`
	Metadata  json.RawMessage `gorm:"type:jsonb"`
	IsUsed    bool            `gorm:"not null;default:false"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`
}

func (f *FileStorage) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	return nil
}