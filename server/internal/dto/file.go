package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateFileData struct {
	URL        string
	TargetID   *string
	IsUsed     bool
	Path       string
	Metadata   json.RawMessage
	Module     string
	Size       int64
	CreatedBy  *string
	FileBuffer []byte 
}

 
type FileResponse struct {
	ID        uuid.UUID `json:"id"`
	URL       string    `json:"url"`
	Path      string    `json:"path"`
	Module    string    `json:"module"`
	IsUsed    bool      `json:"is_used"`
	CreatedAt time.Time `json:"created_at"`
}