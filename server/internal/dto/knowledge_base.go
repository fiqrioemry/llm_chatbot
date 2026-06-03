package dto

import "time"

type CreateKnowledgeBaseRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=255"`
	Description  string `json:"description"`
	EmbedModel   string `json:"embed_model"`
	ChunkSize    int    `json:"chunk_size"`
	ChunkOverlap int    `json:"chunk_overlap"`
}

type UpdateKnowledgeBaseRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description"`
}

type KnowledgeBaseResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	EmbedModel    string    `json:"embed_model"`
	ChunkSize     int       `json:"chunk_size"`
	ChunkOverlap  int       `json:"chunk_overlap"`
	DocumentCount int64     `json:"document_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
