package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

 
type UserStatus string
type UserRole string
type Gender string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
	UserStatusBanned   UserStatus = "BANNED"

	UserRoleUser  UserRole = "USER"
	UserRoleAdmin UserRole = "ADMIN"

	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
)

type User struct {
	ID                   string         `gorm:"type:varchar(36);primaryKey"`
	Email                string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Fullname            string         `gorm:"type:varchar(100);uniqueIndex;not null"`
	PasswordHash         *string        `gorm:"type:varchar(255)"`
	Avatar               *string        `gorm:"type:varchar(500)"`
	Status               UserStatus     `gorm:"type:varchar(20);default:'INACTIVE';not null"`
	Role                 UserRole       `gorm:"type:varchar(20);default:'USER';not null"`
	VerifiedAt           *time.Time
	LastLogin            *time.Time
	LastChangePasswordAt *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            gorm.DeletedAt `gorm:"index"`

	// Relations
	Sessions      []Session      `gorm:"foreignKey:UserID"`
	Verifications []UserVerification `gorm:"foreignKey:UserID"`
	ProviderAccounts []ProviderAccount `gorm:"foreignKey:UserID"`
	ActivityLogs  []UserActivityLog  `gorm:"foreignKey:UserID"`
}


 

type UserVerification struct {
	ID        string     `gorm:"type:varchar(36);primaryKey"`
	UserID    string     `gorm:"type:varchar(36);not null;index"`
	Token     string     `gorm:"type:varchar(255);uniqueIndex;not null"` // SHA256 hashed
	Type      string     `gorm:"type:varchar(50);not null"`              // EMAIL_VERIFICATION | PASSWORD_RESET
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time

	User User `gorm:"foreignKey:UserID"`
}

type OAuthProvider string

const (
	OAuthGoogle OAuthProvider = "GOOGLE"
	OAuthGithub OAuthProvider = "GITHUB"
)

type ProviderAccount struct {
	ID                string        `gorm:"type:varchar(36);primaryKey"`
	UserID            string        `gorm:"type:varchar(36);not null;index"`
	Provider          OAuthProvider `gorm:"type:varchar(50);not null"`
	ProviderAccountID string        `gorm:"type:varchar(255);not null"`
	AccessToken       string        `gorm:"type:text;not null"`
	RefreshToken      *string       `gorm:"type:text"`
	ExpiresAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time

	User User `gorm:"foreignKey:UserID"`
}

type UserActivityLog struct {
	ID        string         `gorm:"type:varchar(36);primaryKey"`
	UserID    string         `gorm:"type:varchar(36);not null;index"`
	Action    string         `gorm:"type:varchar(100);not null"`
	Metadata  string         `gorm:"type:json"`  
	CreatedAt time.Time

	User User `gorm:"foreignKey:UserID"`
}

 

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return
}

func (uv *UserVerification) BeforeCreate(tx *gorm.DB) (err error) {
	if uv.ID == "" {
		uv.ID = uuid.New().String()
	}
	return
}

 
func (oa *ProviderAccount) BeforeCreate(tx *gorm.DB) (err error) {
	if oa.ID == "" {
		oa.ID = uuid.New().String()
	}
	return
}

func (ual *UserActivityLog) BeforeCreate(tx *gorm.DB) (err error) {
	if ual.ID == "" {
		ual.ID = uuid.New().String()
	}
	return
}
 
