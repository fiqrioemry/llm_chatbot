package repositories

import (
	"server/internal/models"
	"server/pkg/pagination"

	pgvector "github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type SimilarChunk struct {
	ChunkID          string
	DocumentID       string
	DocumentFilename string
	Content          string
	SimilarityScore  float32
	Rank             int
}

type ChatRepositoryContract interface {
	CreateConversation(c *models.Conversation) error
	FindConversations(userID string, params pagination.PaginationParams) ([]models.Conversation, int64, error)
	FindConversationByID(id string) (*models.Conversation, error)
	UpdateConversationTitle(id, title string) error
	DeleteConversation(id string) error
	CreateMessage(m *models.Message) error
	FindMessages(conversationID string, params pagination.PaginationParams) ([]models.Message, int64, error)
	FindRecentMessages(conversationID string, limit int) ([]models.Message, error)
	CreateMessageSources(sources []models.MessageSource) error
	FindMessageSources(messageID string) ([]models.MessageSource, error)
	SearchSimilarChunks(kbID string, embedding []float32, topK int) ([]SimilarChunk, error)
}

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateConversation(c *models.Conversation) error {
	return r.db.Create(c).Error
}

func (r *ChatRepository) FindConversations(userID string, params pagination.PaginationParams) ([]models.Conversation, int64, error) {
	var convs []models.Conversation
	var total int64

	q := r.db.Model(&models.Conversation{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Order("updated_at DESC").
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&convs).Error; err != nil {
		return nil, 0, err
	}

	return convs, total, nil
}

func (r *ChatRepository) FindConversationByID(id string) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.Where("id = ?", id).First(&conv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

func (r *ChatRepository) UpdateConversationTitle(id, title string) error {
	return r.db.Model(&models.Conversation{}).Where("id = ?", id).Update("title", title).Error
}

func (r *ChatRepository) DeleteConversation(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Conversation{}).Error
}

func (r *ChatRepository) CreateMessage(m *models.Message) error {
	return r.db.Create(m).Error
}

func (r *ChatRepository) FindMessages(conversationID string, params pagination.PaginationParams) ([]models.Message, int64, error) {
	var msgs []models.Message
	var total int64

	q := r.db.Model(&models.Message{}).Where("conversation_id = ?", conversationID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Order("created_at ASC").
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&msgs).Error; err != nil {
		return nil, 0, err
	}

	return msgs, total, nil
}

func (r *ChatRepository) FindRecentMessages(conversationID string, limit int) ([]models.Message, error) {
	var msgs []models.Message
	err := r.db.Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Find(&msgs).Error
	if err != nil {
		return nil, err
	}

	// Reverse to get chronological order
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (r *ChatRepository) CreateMessageSources(sources []models.MessageSource) error {
	if len(sources) == 0 {
		return nil
	}
	return r.db.Create(&sources).Error
}

func (r *ChatRepository) FindMessageSources(messageID string) ([]models.MessageSource, error) {
	var sources []models.MessageSource
	err := r.db.Preload("Chunk").
		Where("message_id = ?", messageID).
		Order("rank ASC").
		Find(&sources).Error
	return sources, err
}

func (r *ChatRepository) SearchSimilarChunks(kbID string, embedding []float32, topK int) ([]SimilarChunk, error) {
	vec := pgvector.NewVector(embedding)

	type rawRow struct {
		ChunkID          string  `gorm:"column:chunk_id"`
		DocumentID       string  `gorm:"column:document_id"`
		DocumentFilename string  `gorm:"column:document_filename"`
		Content          string  `gorm:"column:content"`
		SimilarityScore  float32 `gorm:"column:similarity_score"`
	}

	var rows []rawRow
	err := r.db.Raw(`
		SELECT
			dc.id AS chunk_id,
			dc.document_id,
			d.filename AS document_filename,
			dc.content,
			CAST(1 - (dc.embedding <=> ?) AS real) AS similarity_score
		FROM document_chunks dc
		JOIN documents d ON d.id = dc.document_id
		WHERE d.knowledge_base_id = ?
		  AND d.status = 'READY'
		  AND dc.embedding IS NOT NULL
		ORDER BY dc.embedding <=> ?
		LIMIT ?
	`, vec, kbID, vec, topK).Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	result := make([]SimilarChunk, len(rows))
	for i, row := range rows {
		result[i] = SimilarChunk{
			ChunkID:          row.ChunkID,
			DocumentID:       row.DocumentID,
			DocumentFilename: row.DocumentFilename,
			Content:          row.Content,
			SimilarityScore:  row.SimilarityScore,
			Rank:             i + 1,
		}
	}

	return result, nil
}
