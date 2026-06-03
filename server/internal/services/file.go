package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"server/internal/config"
	"server/internal/dto"
	"server/internal/repositories"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

const defaultBucket = "uploads"

type FileService struct {
	fileRepo   *repositories.FileRepository
	minio  *minio.Client
	cfg    *config.Config
	bucket string
}

func NewFileService(repo *repositories.FileRepository, minioClient *minio.Client, cfg *config.Config) *FileService {
	bucket := cfg.Minio.Bucket
	if bucket == "" {
		bucket = defaultBucket
	}
	return &FileService{fileRepo: repo, minio: minioClient, cfg: cfg, bucket: bucket}
}

 
func (s *FileService) GenerateFileRecord(fileHeader *multipart.FileHeader, module string, createdBy *string) (*dto.CreateFileData, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("file.FileService.GenerateFileRecord: open: %w", err)
	}
	defer src.Close()

	buf, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("file.Service.GenerateFileRecord: read: %w", err)
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileHeader.Filename), "."))
	if ext == "" {
		ext = "bin"
	}

	randomName := fmt.Sprintf("%s_%d.%s", uuid.NewString(), time.Now().UnixMilli(), ext)
	storageKey := fmt.Sprintf("%s/%s", module, randomName)
	url := fmt.Sprintf("%s/%s/%s", s.cfg.Minio.Endpoint, s.bucket, storageKey)

	meta, _ := json.Marshal(map[string]string{
		"original_name": fileHeader.Filename,
		"mime_type":     fileHeader.Header.Get("Content-Type"),
	})

	return &dto.CreateFileData{
		URL:        url,
		IsUsed:     false,
		Size:       fileHeader.Size,
		Path:       storageKey,
		Metadata:   meta,
		Module:     module,
		CreatedBy:  createdBy,
		FileBuffer: buf,
	}, nil
}

func (s *FileService) UploadToStorage(ctx context.Context, data *dto.CreateFileData) error {
    if s.minio == nil {
        return fmt.Errorf("minio client is nil")
    }
    if len(data.FileBuffer) == 0 {
        return fmt.Errorf("file buffer is empty")
    }
    // guard ctx nil
    if ctx == nil {
        ctx = context.Background()
    }

    contentType := "application/octet-stream"
    var metaMap map[string]string
    if err := json.Unmarshal(data.Metadata, &metaMap); err == nil {
        if mt, ok := metaMap["mime_type"]; ok && mt != "" {
            contentType = mt
        }
    }

    // log sebelum PutObject untuk konfirmasi parameter
    log.Printf("[UploadToStorage] bucket=%s path=%s size=%d contentType=%s bufLen=%d",
        s.bucket, data.Path, data.Size, contentType, len(data.FileBuffer))

    _, err := s.minio.PutObject(
        ctx,
        s.bucket,
        data.Path,
        bytes.NewReader(data.FileBuffer),
        data.Size,
        minio.PutObjectOptions{ContentType: contentType},
    )
    if err != nil {
        log.Printf("[UploadToStorage] error: %v", err)
    }
    return err
}

// SaveRecord menyimpan satu file record ke DB dalam transaksi.
func (s *FileService) SaveRecord(tx *gorm.DB, data *dto.CreateFileData) (*dto.FileResponse, error) {
	return s.fileRepo.CreateFileRecord(tx, *data)
}

// SaveBulkRecords menyimpan banyak file record ke DB dalam transaksi.
func (s *FileService) SaveBulkRecords(tx *gorm.DB, data []*dto.CreateFileData) ([]dto.FileResponse, error) {
	flat := make([]dto.CreateFileData, 0, len(data))
	for _, d := range data {
		flat = append(flat, *d)
	}
	return s.fileRepo.CreateManyFileRecords(tx, flat)
}

// DeleteFile menghapus file dari storage dan DB.
func (s *FileService) DeleteFile(fileID uuid.UUID) error {
	record, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("file not found: %s", fileID)
	}

	if err := s.minio.RemoveObject(nil, s.bucket, record.Path, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("file.Service.DeleteFile: remove object: %w", err)
	}

	return s.fileRepo.DeleteFileRecord(fileID)
}