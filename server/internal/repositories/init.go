package repositories

import "gorm.io/gorm"

type Repository struct {
	Auth *AuthRepository
	User *UserRepository
	File *FileRepository
}


func InitRepository(db *gorm.DB) *Repository {	
	authRepo := NewAuthRepository(db)
	userRepo := NewUserRepository(db)
	fileRepo := NewFileRepository(db)
	return &Repository{
		Auth: authRepo,
		User: userRepo,
		File: fileRepo,
	}
}