package repositories

import (
	"server/internal/dto"
	"server/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) CreateFileRecord(tx *gorm.DB, data dto.CreateFileData) (*dto.FileResponse, error) {
	db := r.resolveDB(tx)

	       var targetID uuid.UUID
	       if data.TargetID != nil {
		       tid, err := uuid.Parse(*data.TargetID)
		       if err != nil {
			       return nil, err
		       }
		       targetID = tid
	       } else {
		       targetID = uuid.Nil
	       }

	       record := &models.FileStorage{
		       ID:        uuid.New().String(),
		       TargetID:  targetID.String(),
		       Module:    data.Module,
		       URL:       data.URL,
		       Path:      data.Path,
		       Metadata:  data.Metadata,
		       IsUsed:    data.IsUsed,
	       }

	if err := db.Create(record).Error; err != nil {
		return nil, err
	}

	return toFileResponse(record), nil
}

func (r *FileRepository) CreateManyFileRecords(tx *gorm.DB, data []dto.CreateFileData) ([]dto.FileResponse, error) {
	db := r.resolveDB(tx)

	records := make([]models.FileStorage, 0, len(data))
	for _, d := range data {

	       var targetID uuid.UUID
	       if d.TargetID != nil {
		       tid, err := uuid.Parse(*d.TargetID)
		       if err != nil {
			       return nil, err
		       }
		       targetID = tid
	       } else {
		       targetID = uuid.Nil
	       }

		records = append(records, models.FileStorage{
			ID:        uuid.New().String(),
			TargetID:  targetID.String(),
			Module:    d.Module,
			URL:       d.URL,
			Path:      d.Path,
			Metadata:  d.Metadata,
			IsUsed:    d.IsUsed,
		})
	}

	if err := db.Create(&records).Error; err != nil {
		return nil, err
	}

	responses := make([]dto.FileResponse, 0, len(records))
	for i := range records {
		responses = append(responses, *toFileResponse(&records[i]))
	}

	return responses, nil
}

func (r *FileRepository) GetFileByID(fileID uuid.UUID) (*dto.FileResponse, error) {
	var record models.FileStorage
	err := r.db.Where("id = ?", fileID).First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toFileResponse(&record), nil
}

func (r *FileRepository) GetFileByTargetID(targetID, module string) (*dto.FileResponse, error) {
	var record models.FileStorage
	err := r.db.Where("target_id = ? AND module = ?", targetID, module).First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toFileResponse(&record), nil
}

func (r *FileRepository) DeleteFileRecord(fileID uuid.UUID) error {
	return r.db.Where("id = ?", fileID).Delete(&models.FileStorage{}).Error
}

func (r *FileRepository) MarkFilesAsUsed(tx *gorm.DB, ids []uuid.UUID) error {
	db := r.resolveDB(tx)
	return db.Model(&models.FileStorage{}).
		Where("id IN ?", ids).
		Update("is_used", true).Error
}

func (r *FileRepository) MarkFilesAsUnused(tx *gorm.DB, ids []uuid.UUID) error {
	db := r.resolveDB(tx)
	return db.Model(&models.FileStorage{}).
		Where("id IN ?", ids).
		Update("is_used", false).Error
}

func (r *FileRepository) UpdateFileTargetIDs(tx *gorm.DB, updates []struct {
	FileID   uuid.UUID
	TargetID string
}) error {
	db := r.resolveDB(tx)
	for _, u := range updates {
		if err := db.Model(&models.FileStorage{}).
			Where("id = ?", u.FileID).
			Updates(map[string]interface{}{
				"target_id": u.TargetID,
				"is_used":   true,
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *FileRepository) FindExpiredUnusedFiles(ttl time.Duration) ([]models.FileStorage, error) {
	var records []models.FileStorage
	cutoff := time.Now().Add(-ttl)
	err := r.db.Where("is_used = false AND created_at < ?", cutoff).
		Select("id, path").
		Find(&records).Error
	return records, err
}

func (r *FileRepository) DeleteFileRecordsByIDs(ids []uuid.UUID) (int64, error) {
	result := r.db.Where("id IN ?", ids).Delete(&models.FileStorage{})
	return result.RowsAffected, result.Error
}

// resolveDB mengembalikan tx jika tidak nil, jika tidak menggunakan db utama
func (r *FileRepository) resolveDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func toFileResponse(f *models.FileStorage) *dto.FileResponse {
	return &dto.FileResponse{
		ID:        f.ID,
		URL:       f.URL,
		Path:      f.Path,
		Module:    f.Module,
		IsUsed:    f.IsUsed,
		CreatedAt: f.CreatedAt,
	}
}