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
	ID                   string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email                string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Fullname             string         `gorm:"type:varchar(100);not null"`
	PasswordHash         *string        `gorm:"type:varchar(255)"`
	Avatar               *string        `gorm:"type:varchar(500)"`
	Status               UserStatus     `gorm:"type:varchar(20);default:'INACTIVE';not null"`
	Role                 UserRole       `gorm:"type:varchar(20);default:'USER';not null"`
	VerifiedAt           *time.Time
	LastLogin            *time.Time
	LastChangePasswordAt *time.Time
	CreatedAt            time.Time      `gorm:"autoCreateTime"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime"`
	DeletedAt            gorm.DeletedAt `gorm:"index"`

	// Relations
	Sessions         []Session         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Verifications    []UserVerification `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	ProviderAccounts []ProviderAccount  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	ActivityLogs     []UserActivityLog  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Conversations    []Conversation     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type UserVerification struct {
	ID        string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string     `gorm:"type:uuid;not null;index"`
	Token     string     `gorm:"type:varchar(255);uniqueIndex;not null"` // SHA256 hashed
	Type      string     `gorm:"type:varchar(50);not null"`              // EMAIL_VERIFICATION | PASSWORD_RESET
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time  `gorm:"autoCreateTime"`

	User User `gorm:"foreignKey:UserID"`
}

type OAuthProvider string

const (
	OAuthGoogle OAuthProvider = "GOOGLE"
	OAuthGithub OAuthProvider = "GITHUB"
)

type ProviderAccount struct {
	ID                string        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            string        `gorm:"type:uuid;not null;index"`
	Provider          OAuthProvider `gorm:"type:varchar(50);not null"`
	ProviderAccountID string        `gorm:"type:varchar(255);not null"`
	AccessToken       string        `gorm:"type:text;not null"`
	RefreshToken      *string       `gorm:"type:text"`
	ExpiresAt         *time.Time
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`

	User User `gorm:"foreignKey:UserID"`
}

func (pa *ProviderAccount) TableName() string {
	return "provider_accounts"
}

type UserActivityLog struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	Action    string    `gorm:"type:varchar(100);not null"`
	Metadata  []byte    `gorm:"type:jsonb"`
	IPAddress string    `gorm:"type:varchar(50)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	User User `gorm:"foreignKey:UserID"`
}

// Hooks

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

func (uv *UserVerification) BeforeCreate(tx *gorm.DB) error {
	if uv.ID == "" {
		uv.ID = uuid.New().String()
	}
	return nil
}

func (oa *ProviderAccount) BeforeCreate(tx *gorm.DB) error {
	if oa.ID == "" {
		oa.ID = uuid.New().String()
	}
	return nil
}

func (ual *UserActivityLog) BeforeCreate(tx *gorm.DB) error {
	if ual.ID == "" {
		ual.ID = uuid.New().String()
	}
	return nil
}