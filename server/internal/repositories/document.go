package repositories

import (
	"server/internal/models"
	"server/pkg/pagination"

	"gorm.io/gorm"
)

type DocumentRepositoryContract interface {
	Create(doc *models.Document) error
	FindAll(kbID string, status string, params pagination.PaginationParams) ([]models.Document, int64, error)
	FindByID(id string) (*models.Document, error)
	UpdateStatus(id string, status models.DocumentStatus, errMsg *string) error
	Delete(id string) error
	CountChunks(docID string) (int64, error)
	CreateChunks(chunks []models.DocumentChunk) error
	DeleteChunksByDocumentID(docID string) error
}

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(doc *models.Document) error {
	return r.db.Create(doc).Error
}

func (r *DocumentRepository) FindAll(kbID string, status string, params pagination.PaginationParams) ([]models.Document, int64, error) {
	var docs []models.Document
	var total int64

	q := r.db.Model(&models.Document{}).Where("knowledge_base_id = ?", kbID)
	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Order("created_at DESC").
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&docs).Error; err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

func (r *DocumentRepository) FindByID(id string) (*models.Document, error) {
	var doc models.Document
	err := r.db.Where("id = ?", id).First(&doc).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &doc, nil
}

func (r *DocumentRepository) UpdateStatus(id string, status models.DocumentStatus, errMsg *string) error {
	updates := map[string]any{"status": status}
	if errMsg != nil {
		updates["error_message"] = errMsg
	}
	return r.db.Model(&models.Document{}).Where("id = ?", id).Updates(updates).Error
}

func (r *DocumentRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Document{}).Error
}

func (r *DocumentRepository) CountChunks(docID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.DocumentChunk{}).Where("document_id = ?", docID).Count(&count).Error
	return count, err
}

func (r *DocumentRepository) CreateChunks(chunks []models.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	return r.db.Create(&chunks).Error
}

func (r *DocumentRepository) DeleteChunksByDocumentID(docID string) error {
	return r.db.Where("document_id = ?", docID).Delete(&models.DocumentChunk{}).Error
}
