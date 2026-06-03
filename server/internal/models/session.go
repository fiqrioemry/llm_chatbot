package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       string    `gorm:"type:uuid;not null;index"`
	RefreshToken string    `gorm:"type:varchar(255);uniqueIndex;not null"` // SHA256 hashed
	UserAgent    string    `gorm:"type:varchar(500)"`
	IPAddress    string    `gorm:"type:varchar(50)"`
	ExpiresAt    time.Time `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}