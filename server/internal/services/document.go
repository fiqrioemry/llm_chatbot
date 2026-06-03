package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"unicode"

	"server/internal/config"
	"server/internal/config/constant"
	"server/internal/dto"
	"server/internal/models"
	"server/internal/repositories"
	"server/pkg/pagination"
	"server/pkg/response"
	"server/pkg/utils"
	"server/providers/openai"

	pgvector "github.com/pgvector/pgvector-go"
	"github.com/google/uuid"
	miniogo "github.com/minio/minio-go/v7"
)

const maxDocumentSize = 20 * 1024 * 1024 // 20MB

var allowedFileTypes = map[string]bool{
	"pdf":  true,
	"txt":  true,
	"md":   true,
	"docx": true,
}

type DocumentService struct {
	docRepo repositories.DocumentRepositoryContract
	kbRepo  repositories.KnowledgeBaseRepositoryContract
	ai      *openai.Client
	minio   *miniogo.Client
	cfg     *config.Config
	bucket  string
}

func NewDocumentService(
	docRepo repositories.DocumentRepositoryContract,
	kbRepo repositories.KnowledgeBaseRepositoryContract,
	ai *openai.Client,
	minio *miniogo.Client,
	cfg *config.Config,
) *DocumentService {
	return &DocumentService{
		docRepo: docRepo,
		kbRepo:  kbRepo,
		ai:      ai,
		minio:   minio,
		cfg:     cfg,
		bucket:  cfg.Minio.Bucket,
	}
}

func (s *DocumentService) Upload(ctx context.Context, kbID string, fileHeader *multipart.FileHeader) (*dto.DocumentResponse, error) {
	// Validate KB exists
	kb, err := s.kbRepo.FindByID(kbID)
	if err != nil {
		return nil, response.InternalErr(err.Error(), "KB_FETCH_FAIL")
	}
	if kb == nil {
		return nil, response.NotFoundErr(constant.ErrKnowledgeBaseNotFound, constant.CodeKnowledgeBaseNotFound)
	}

	// Validate file size
	if fileHeader.Size > maxDocumentSize {
		return nil, response.BadRequestErr(constant.ErrDocumentTooLarge, constant.CodeDocumentTooLarge)
	}

	// Validate file type
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileHeader.Filename), "."))
	if !allowedFileTypes[ext] {
		return nil, response.BadRequestErr(constant.ErrDocumentInvalidType, constant.CodeDocumentInvalidType)
	}

	// Generate document ID
	docID := uuid.New().String()
	objectPath := fmt.Sprintf("documents/%s/%s/%s", kbID, docID, fileHeader.Filename)

	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, response.InternalErr("failed to open uploaded file", "FILE_OPEN_FAIL")
	}
	defer file.Close()

	// Upload to MinIO
	_, err = s.minio.PutObject(ctx, s.bucket, objectPath, file, fileHeader.Size, miniogo.PutObjectOptions{
		ContentType: fileHeader.Header.Get("Content-Type"),
	})
	if err != nil {
		return nil, response.InternalErr(fmt.Sprintf("failed to upload to storage: %v", err), "MINIO_UPLOAD_FAIL")
	}

	// Create document record
	doc := models.Document{
		ID:              docID,
		KnowledgeBaseID: kbID,
		Filename:        fileHeader.Filename,
		FileType:        ext,
		FileSize:        fileHeader.Size,
		Status:          models.DocumentStatusPending,
	}

	if err := s.docRepo.Create(&doc); err != nil {
		return nil, response.InternalErr(err.Error(), "DOC_CREATE_FAIL")
	}

	log.Printf("[document service] Upload — doc created: %s, kb: %s", doc.ID, kbID)

	// Start ingestion asynchronously
	go s.ingest(doc, *kb, objectPath)

	res := s.toResponse(doc, 0)
	return &res, nil
}

func (s *DocumentService) ingest(doc models.Document, kb models.KnowledgeBase, objectPath string) {
	ctx := context.Background()
	log.Printf("[worker:ingest] started — doc: %s", doc.ID)

	// Mark as processing
	if err := s.docRepo.UpdateStatus(doc.ID, models.DocumentStatusProcessing, nil); err != nil {
		log.Printf("[worker:ingest] failed — doc: %s, error: %v", doc.ID, err)
		return
	}

	// Get file from MinIO
	obj, err := s.minio.GetObject(ctx, s.bucket, objectPath, miniogo.GetObjectOptions{})
	if err != nil {
		errMsg := fmt.Sprintf("failed to get file from storage: %v", err)
		log.Printf("[worker:ingest] failed — doc: %s, error: %v", doc.ID, err)
		s.docRepo.UpdateStatus(doc.ID, models.DocumentStatusFailed, &errMsg)
		return
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read file: %v", err)
		log.Printf("[worker:ingest] failed — doc: %s, error: %v", doc.ID, err)
		s.docRepo.UpdateStatus(doc.ID, models.DocumentStatusFailed, &errMsg)
		return
	}

	// Extract text
	var text string
	switch doc.FileType {
	case "txt", "md":
		text = string(data)
	case "pdf":
		text = extractPDFText(data)
	case "docx":
		text, err = extractDocxText(data)
		if err != nil {
			errMsg := fmt.Sprintf("failed to extract docx text: %v", err)
			log.Printf("[worker:ingest] failed — doc: %s, error: %v", doc.ID, err)
			s.docRepo.UpdateStatus(doc.ID, models.DocumentStatusFailed, &errMsg)
			return
		}
	default:
		errMsg := fmt.Sprintf("unsupported file type: %s", doc.FileType)
		log.Printf("[worker:ingest] failed — doc: %s, error: %s", doc.ID, errMsg)
		s.docRepo.UpdateStatus(doc.ID, models.DocumentStatusFailed, &errMsg)
		return
	}

	// Chunk text
	chunkSize := kb.ChunkSize
	if chunkSize == 0 {
		chunkSize = 512
	}
	chunkOverlap := kb.ChunkOverlap

	chunks := utils.ChunkText(text, chunkSize, chunkOverlap)
	log.Printf("[worker:ingest] chunked %d chunks — doc: %s", len(chunks), doc.ID)

	if len(chunks) == 0 {
		errMsg := "no text content extracted from document"
		log.Printf("[worker:ingest] failed — doc: %s, error: %s", doc.ID, errMsg)
		s.docRepo.UpdateStatus(doc.ID, models.DocumentStatusFailed, &errMsg)
		return
	}

	// Process in batches of 20
	batchSize := 20
	totalChunks := 0
	totalBatches := (len(chunks) + batchSize - 1) / batchSize

	for batchNum := 0; batchNum < totalBatches; batchNum++ {
		start := batchNum * batchSize
		end := start + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		batchTexts := chunks[start:end]

		vectors, err := s.ai.EmbedBatch(ctx, batchTexts)
		if err != nil {
			errMsg := fmt.Sprintf("embedding failed: %v", err)
			log.Printf("[worker:ingest] failed — doc: %s, error: %v", doc.ID, err)
			s.docRepo.UpdateStatus(doc.ID, models.DocumentStatusFailed, &errMsg)
			return
		}

		log.Printf("[worker:ingest] embedded batch %d/%d — doc: %s", batchNum+1, totalBatches, doc.ID)

		docChunks := make([]models.DocumentChunk, len(batchTexts))
		for i, content := range batchTexts {
			docChunks[i] = models.DocumentChunk{
				DocumentID: doc.ID,
				ChunkIndex: start + i,
				Content:    content,
				Embedding:  pgvector.NewVector(vectors[i]),
				Metadata:   nil,
			}
		}

		if err := s.docRepo.CreateChunks(docChunks); err != nil {
			errMsg := fmt.Sprintf("failed to save chunks: %v", err)
			log.Printf("[worker:ingest] failed — doc: %s, error: %v", doc.ID, err)
			s.docRepo.UpdateStatus(doc.ID, models.DocumentStatusFailed, &errMsg)
			return
		}

		totalChunks += len(batchTexts)
	}

	// Mark as ready
	if err := s.docRepo.UpdateStatus(doc.ID, models.DocumentStatusReady, nil); err != nil {
		log.Printf("[worker:ingest] failed to mark ready — doc: %s, error: %v", doc.ID, err)
		return
	}

	log.Printf("[worker:ingest] completed — doc: %s, chunks: %d", doc.ID, totalChunks)
}

func (s *DocumentService) List(kbID, status string, params pagination.PaginationParams) ([]dto.DocumentResponse, int64, error) {
	docs, total, err := s.docRepo.FindAll(kbID, status, params)
	if err != nil {
		return nil, 0, response.InternalErr(err.Error(), "DOC_LIST_FAIL")
	}

	result := make([]dto.DocumentResponse, 0, len(docs))
	for _, doc := range docs {
		count, _ := s.docRepo.CountChunks(doc.ID)
		result = append(result, s.toResponse(doc, count))
	}

	return result, total, nil
}

func (s *DocumentService) GetByID(kbID, docID string) (*dto.DocumentResponse, error) {
	doc, err := s.docRepo.FindByID(docID)
	if err != nil {
		return nil, response.InternalErr(err.Error(), "DOC_FETCH_FAIL")
	}
	if doc == nil || doc.KnowledgeBaseID != kbID {
		return nil, response.NotFoundErr(constant.ErrDocumentNotFound, constant.CodeDocumentNotFound)
	}

	count, _ := s.docRepo.CountChunks(doc.ID)
	res := s.toResponse(*doc, count)
	return &res, nil
}

func (s *DocumentService) Delete(ctx context.Context, kbID, docID string) error {
	doc, err := s.docRepo.FindByID(docID)
	if err != nil {
		return response.InternalErr(err.Error(), "DOC_FETCH_FAIL")
	}
	if doc == nil || doc.KnowledgeBaseID != kbID {
		return response.NotFoundErr(constant.ErrDocumentNotFound, constant.CodeDocumentNotFound)
	}

	// Remove from MinIO
	objectPath := fmt.Sprintf("documents/%s/%s/%s", kbID, docID, doc.Filename)
	_ = s.minio.RemoveObject(ctx, s.bucket, objectPath, miniogo.RemoveObjectOptions{})

	// Delete chunks explicitly
	if err := s.docRepo.DeleteChunksByDocumentID(docID); err != nil {
		return response.InternalErr(err.Error(), "DOC_CHUNK_DELETE_FAIL")
	}

	// Delete document
	if err := s.docRepo.Delete(docID); err != nil {
		return response.InternalErr(err.Error(), "DOC_DELETE_FAIL")
	}

	return nil
}

func (s *DocumentService) toResponse(doc models.Document, chunkCount int64) dto.DocumentResponse {
	return dto.DocumentResponse{
		ID:              doc.ID,
		KnowledgeBaseID: doc.KnowledgeBaseID,
		Filename:        doc.Filename,
		FileType:        doc.FileType,
		FileSize:        doc.FileSize,
		Status:          doc.Status,
		ErrorMessage:    doc.ErrorMessage,
		ChunkCount:      chunkCount,
		CreatedAt:       doc.CreatedAt,
		UpdatedAt:       doc.UpdatedAt,
	}
}

// extractPDFText scans raw PDF bytes and collects printable ASCII runs of length > 3.
func extractPDFText(data []byte) string {
	var sb strings.Builder
	var run strings.Builder

	flush := func() {
		if run.Len() > 3 {
			if sb.Len() > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(run.String())
		}
		run.Reset()
	}

	for _, b := range data {
		if b >= 32 && b < 127 && unicode.IsPrint(rune(b)) {
			run.WriteByte(b)
		} else {
			flush()
		}
	}
	flush()

	return sb.String()
}

// extractDocxText unzips a docx file and extracts text from word/document.xml.
func extractDocxText(data []byte) (string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to open docx as zip: %w", err)
	}

	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open word/document.xml: %w", err)
		}
		defer rc.Close()

		xmlData, err := io.ReadAll(rc)
		if err != nil {
			return "", fmt.Errorf("failed to read word/document.xml: %w", err)
		}

		return stripXMLTags(xmlData), nil
	}

	return "", fmt.Errorf("word/document.xml not found in docx")
}

// stripXMLTags decodes XML and collects character data, joining with spaces.
func stripXMLTags(data []byte) string {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var sb strings.Builder

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		if cd, ok := tok.(xml.CharData); ok {
			text := strings.TrimSpace(string(cd))
			if text != "" {
				if sb.Len() > 0 {
					sb.WriteByte(' ')
				}
				sb.WriteString(text)
			}
		}
	}

	return sb.String()
}
