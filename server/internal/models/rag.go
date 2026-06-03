package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// ─── Enums ───────────────────────────────────────────────────────────────────

type DocumentStatus string
type MessageRole string

const (
	DocumentStatusPending    DocumentStatus = "PENDING"
	DocumentStatusProcessing DocumentStatus = "PROCESSING"
	DocumentStatusReady      DocumentStatus = "READY"
	DocumentStatusFailed     DocumentStatus = "FAILED"

	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

// ─── KnowledgeBase ───────────────────────────────────────────────────────────

type KnowledgeBase struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string    `gorm:"type:varchar(255);not null"`
	Description  string    `gorm:"type:text"`
	EmbedModel   string    `gorm:"type:varchar(100);not null;default:'text-embedding-3-small'"`
	ChunkSize    int       `gorm:"not null;default:512"`
	ChunkOverlap int       `gorm:"not null;default:64"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`

	// Relations
	Documents     []Document     `gorm:"foreignKey:KnowledgeBaseID;constraint:OnDelete:CASCADE"`
	Conversations []Conversation `gorm:"foreignKey:KnowledgeBaseID"`
}

// ─── Document ─────────────────────────────────────────────────────────────────

type Document struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	KnowledgeBaseID string         `gorm:"type:uuid;not null;index"`
	Filename        string         `gorm:"type:varchar(255);not null"`
	FileType        string         `gorm:"type:varchar(50);not null"` // pdf, txt, md, docx
	FileSize        int64          `gorm:"not null"`
	Status          DocumentStatus `gorm:"type:varchar(20);not null;default:'PENDING'"`
	ErrorMessage    *string        `gorm:"type:text"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`

	// Relations
	KnowledgeBase KnowledgeBase   `gorm:"foreignKey:KnowledgeBaseID"`
	Chunks        []DocumentChunk `gorm:"foreignKey:DocumentID;constraint:OnDelete:CASCADE"`
}

type ChunkMetadata struct {
	Page    *int    `json:"page,omitempty"`
	Section *string `json:"section,omitempty"`
	Header  *string `json:"header,omitempty"`
}

type DocumentChunk struct {
	ID         string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DocumentID string          `gorm:"type:uuid;not null;index"`
	ChunkIndex int             `gorm:"not null"`
	Content    string          `gorm:"type:text;not null"`
	Embedding  pgvector.Vector `gorm:"type:vector(1536)"`  // text-embedding-3-small = 1536 dims
	Metadata   json.RawMessage `gorm:"type:jsonb"`
	CreatedAt  time.Time       `gorm:"autoCreateTime"`

	// Relations
	Document Document        `gorm:"foreignKey:DocumentID"`
	Sources  []MessageSource `gorm:"foreignKey:ChunkID"`
}

func (dc *DocumentChunk) TableName() string {
	return "document_chunks"
}

// ─── Conversation ─────────────────────────────────────────────────────────────

type Conversation struct {
	ID              string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          string    `gorm:"type:uuid;not null;index"`
	KnowledgeBaseID string    `gorm:"type:uuid;not null;index"`
	Title           string    `gorm:"type:varchar(255);not null;default:'New Conversation'"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`

	// Relations
	User          User          `gorm:"foreignKey:UserID"`
	KnowledgeBase KnowledgeBase `gorm:"foreignKey:KnowledgeBaseID"`
	Messages      []Message     `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE"`
}

// ─── Message ──────────────────────────────────────────────────────────────────

type Message struct {
	ID             string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID string      `gorm:"type:uuid;not null;index"`
	Role           MessageRole `gorm:"type:varchar(20);not null"`
	Content        string      `gorm:"type:text;not null"`
	TokensUsed     int         `gorm:"default:0"`
	LatencyMs      int         `gorm:"default:0"`
	CreatedAt      time.Time   `gorm:"autoCreateTime"`

	// Relations
	Conversation Conversation    `gorm:"foreignKey:ConversationID"`
	Sources      []MessageSource `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`
}

// ─── MessageSource ────────────────────────────────────────────────────────────
// Junction table: records which chunks were used as context for an assistant message.

type MessageSource struct {
	ID              string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MessageID       string  `gorm:"type:uuid;not null;index"`
	ChunkID         string  `gorm:"type:uuid;not null;index"`
	SimilarityScore float32 `gorm:"type:real;not null"`
	Rank            int     `gorm:"not null"` // 1 = most relevant

	// Relations
	Message Message       `gorm:"foreignKey:MessageID"`
	Chunk   DocumentChunk `gorm:"foreignKey:ChunkID"`
}

func (ms *MessageSource) TableName() string {
	return "message_sources"
}

// ─── Hooks ────────────────────────────────────────────────────────────────────

func (kb *KnowledgeBase) BeforeCreate(tx *gorm.DB) error {
	if kb.ID == "" {
		kb.ID = uuid.New().String()
	}
	return nil
}

func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

func (dc *DocumentChunk) BeforeCreate(tx *gorm.DB) error {
	if dc.ID == "" {
		dc.ID = uuid.New().String()
	}
	return nil
}

func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

func (ms *MessageSource) BeforeCreate(tx *gorm.DB) error {
	if ms.ID == "" {
		ms.ID = uuid.New().String()
	}
	return nil
}