package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type FileStorage struct {
	ID        uuid.UUID      `gorm:"type:char(36);primaryKey"`
	TargetID uuid.UUID      `gorm:"type:char(36);not null;index"`
	Module  string         `gorm:"type:varchar(50);not null;index"`
	URL     string         `gorm:"type:varchar(255);not null"`
	Path	string         `gorm:"type:varchar(255);not null"`
	Metadata json.RawMessage `gorm:"type:json"`
	IsUsed  bool           `gorm:"not null;default:false"`
	CreatedAt time.Time     
}

func (f *FileStorage) BeforeCreate() error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}