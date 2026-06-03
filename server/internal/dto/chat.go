package dto

import (
	"time"

	"server/internal/models"
)

type CreateConversationRequest struct {
	KnowledgeBaseID string `json:"knowledge_base_id" binding:"required,uuid"`
}

type ConversationResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	KnowledgeBaseID string    `json:"knowledge_base_id"`
	Title           string    `json:"title"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=4000"`
}

type SourceResponse struct {
	ChunkID          string  `json:"chunk_id"`
	Content          string  `json:"content"`
	SimilarityScore  float32 `json:"similarity_score"`
	Rank             int     `json:"rank"`
	DocumentFilename string  `json:"document_filename"`
}

type MessageResponse struct {
	ID             string             `json:"id"`
	ConversationID string             `json:"conversation_id"`
	Role           models.MessageRole `json:"role"`
	Content        string             `json:"content"`
	TokensUsed     int                `json:"tokens_used"`
	LatencyMs      int                `json:"latency_ms"`
	Sources        []SourceResponse   `json:"sources,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
}
