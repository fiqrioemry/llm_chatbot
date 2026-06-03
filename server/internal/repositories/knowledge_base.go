package repositories

import (
	"server/internal/models"
	"server/pkg/pagination"

	"gorm.io/gorm"
)

type KnowledgeBaseRepositoryContract interface {
	Create(kb *models.KnowledgeBase) error
	FindAll(params pagination.PaginationParams) ([]models.KnowledgeBase, int64, error)
	FindByID(id string) (*models.KnowledgeBase, error)
	Update(id string, updates map[string]any) error
	Delete(id string) error
	CountDocuments(kbID string) (int64, error)
}

type KnowledgeBaseRepository struct {
	db *gorm.DB
}

func NewKnowledgeBaseRepository(db *gorm.DB) *KnowledgeBaseRepository {
	return &KnowledgeBaseRepository{db: db}
}

func (r *KnowledgeBaseRepository) Create(kb *models.KnowledgeBase) error {
	return r.db.Create(kb).Error
}

func (r *KnowledgeBaseRepository) FindAll(params pagination.PaginationParams) ([]models.KnowledgeBase, int64, error) {
	var kbs []models.KnowledgeBase
	var total int64

	if err := r.db.Model(&models.KnowledgeBase{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Order("created_at DESC").
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&kbs).Error; err != nil {
		return nil, 0, err
	}

	return kbs, total, nil
}

func (r *KnowledgeBaseRepository) FindByID(id string) (*models.KnowledgeBase, error) {
	var kb models.KnowledgeBase
	err := r.db.Where("id = ?", id).First(&kb).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &kb, nil
}

func (r *KnowledgeBaseRepository) Update(id string, updates map[string]any) error {
	return r.db.Model(&models.KnowledgeBase{}).Where("id = ?", id).Updates(updates).Error
}

func (r *KnowledgeBaseRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.KnowledgeBase{}).Error
}

func (r *KnowledgeBaseRepository) CountDocuments(kbID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Document{}).Where("knowledge_base_id = ?", kbID).Count(&count).Error
	return count, err
}
