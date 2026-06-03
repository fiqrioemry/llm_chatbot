package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID           string    `gorm:"type:varchar(36);primaryKey"`
	UserID       string    `gorm:"type:varchar(36);not null;index"`
	RefreshToken string    `gorm:"type:varchar(255);uniqueIndex;not null"` // SHA256 hashed
	UserAgent    string    `gorm:"type:varchar(500)"`
	IPAddress    string    `gorm:"type:varchar(50)"`
	ExpiresAt    time.Time `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
 
	User User `gorm:"foreignKey:UserID"`
}
 
func (t *Session) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return
}