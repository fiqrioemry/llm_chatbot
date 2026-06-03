package repositories

import "gorm.io/gorm"

type Repository struct {
	Auth          *AuthRepository
	User          *UserRepository
	File          *FileRepository
	KnowledgeBase *KnowledgeBaseRepository
	Document      *DocumentRepository
	Chat          *ChatRepository
}

func InitRepository(db *gorm.DB) *Repository {
	authRepo          := NewAuthRepository(db)
	userRepo          := NewUserRepository(db)
	fileRepo          := NewFileRepository(db)
	knowledgeBaseRepo := NewKnowledgeBaseRepository(db)
	documentRepo      := NewDocumentRepository(db)
	chatRepo          := NewChatRepository(db)

	return &Repository{
		Auth:          authRepo,
		User:          userRepo,
		File:          fileRepo,
		KnowledgeBase: knowledgeBaseRepo,
		Document:      documentRepo,
		Chat:          chatRepo,
	}
}