package dto

import (
	"time"

	"server/internal/models"
)

type DocumentResponse struct {
	ID              string                `json:"id"`
	KnowledgeBaseID string                `json:"knowledge_base_id"`
	Filename        string                `json:"filename"`
	FileType        string                `json:"file_type"`
	FileSize        int64                 `json:"file_size"`
	Status          models.DocumentStatus `json:"status"`
	ErrorMessage    *string               `json:"error_message,omitempty"`
	ChunkCount      int64                 `json:"chunk_count"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}
