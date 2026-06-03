package repositories

import (
	"server/internal/models"

	"gorm.io/gorm"
)

type UserRepositoryContract interface {
	UpdateProfile(tx *gorm.DB, userID string, user *models.User) error
	CreateActivityLog(tx *gorm.DB, log *models.UserActivityLog) error
	GetByID(userID string) (error, *models.User)
	UpdateAvatar(tx *gorm.DB, userID string, url string) error
	Transaction(fn func(tx *gorm.DB) error) error
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) UpdateProfile(tx *gorm.DB, userID string, user *models.User) error {
	return tx.Model(&models.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(user).Error
}	

func (r *UserRepository) CreateActivityLog(tx *gorm.DB, log *models.UserActivityLog) error {
	return tx.Create(log).Error
}


func (r *UserRepository) GetByID(userID string) (error, *models.User) {
	var user models.User
	err := r.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return err, &user
}

func (r *UserRepository) UpdateAvatar(tx *gorm.DB, userID string, url string) error {
	return tx.Model(&models.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("avatar", url).Error
}


func (r *UserRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
