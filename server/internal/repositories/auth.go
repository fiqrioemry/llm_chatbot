package repositories

import (
	"server/internal/models"
	"time"

	"gorm.io/gorm"
)

 type AuthRepositoryContract interface {
	FindUserByEmail(email string) (*models.User, error)
	FindUserByID(id string) (*models.User, error)
	FindUserByOAuthProvider(provider models.OAuthProvider, providerAccountID string) (*models.User, error)
	CreateUser(tx *gorm.DB, user *models.User) error
	UpdateLastLogin(tx *gorm.DB, userID string) error
	UpdateUserStatus(tx *gorm.DB, userID string, status models.UserStatus) error
	UpdatePassword(tx *gorm.DB, userID, hashedPassword string) error
	GetPasswordHash(userID string) (string, error)
	CreateVerification(tx *gorm.DB, v *models.UserVerification) error
	FindVerification(hashedToken, vType string) (*models.UserVerification, error)
	MarkVerificationUsed(tx *gorm.DB, id string) error
	CreateSession(tx *gorm.DB, s *models.Session) error
	FindSessionByRefreshToken(hashedToken string) (*models.Session, error)
	FindSessionByID(id string) (*models.Session, error)
	FindSessionsByUserID(userID string) ([]models.Session, error)
	DeleteSessionByID(tx *gorm.DB, id string) error
	DeleteAllSessionsByUserID(tx *gorm.DB, userID string) error
	UpsertOAuthAccount(tx *gorm.DB, account *models.ProviderAccount) error
	CreateOAuthUser(tx *gorm.DB, user *models.User) error
	Transaction(fn func(tx *gorm.DB) error) error
}

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

 
func (r *AuthRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *AuthRepository) FindUserByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *AuthRepository) FindUserByOAuthProvider(provider models.OAuthProvider, providerAccountID string) (*models.User, error) {
	var account models.ProviderAccount
	err := r.db.
		Preload("User").
		Where("provider = ? AND provider_account_id = ?", provider, providerAccountID).
		First(&account).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &account.User, nil
}

func (r *AuthRepository) CreateUser(tx *gorm.DB, user *models.User) error {
	return tx.Create(user).Error
}

func (r *AuthRepository) UpdateLastLogin(tx *gorm.DB, userID string) error {
	now := time.Now()
	return tx.Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login", now).Error
}

func (r *AuthRepository) UpdateUserStatus(tx *gorm.DB, userID string, status models.UserStatus) error {
	updates := map[string]interface{}{
		"status":      status,
		"verified_at": time.Now(),
	}
	return tx.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *AuthRepository) UpdatePassword(tx *gorm.DB, userID, hashedPassword string) error {
	now := time.Now()
	return tx.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password_hash":           hashedPassword,
			"last_change_password_at": now,
		}).Error
}

func (r *AuthRepository) GetPasswordHash(userID string) (string, error) {
	var user models.User
	err := r.db.Select("password_hash").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return "", err
	}
	if user.PasswordHash == nil {
		return "", nil
	}
	return *user.PasswordHash, nil
}

 
func (r *AuthRepository) CreateVerification(tx *gorm.DB, v *models.UserVerification) error {
	return tx.Create(v).Error
}

func (r *AuthRepository) FindVerification(hashedToken, vType string) (*models.UserVerification, error) {
	var v models.UserVerification
	err := r.db.
		Where("token = ? AND type = ? AND used_at IS NULL AND expires_at > ?", hashedToken, vType, time.Now()).
		First(&v).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &v, err
}

func (r *AuthRepository) MarkVerificationUsed(tx *gorm.DB, id string) error {
	now := time.Now()
	return tx.Model(&models.UserVerification{}).
		Where("id = ?", id).
		Update("used_at", now).Error
}

 
func (r *AuthRepository) CreateSession(tx *gorm.DB, s *models.Session) error {
	return tx.Create(s).Error
}

func (r *AuthRepository) FindSessionByRefreshToken(hashedToken string) (*models.Session, error) {
	var s models.Session
	err := r.db.
		Preload("User").
		Where("refresh_token = ? AND expires_at > ?", hashedToken, time.Now()).
		First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &s, err
}

func (r *AuthRepository) FindSessionByID(id string) (*models.Session, error) {
	var s models.Session
	err := r.db.Where("id = ?", id).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &s, err
}

func (r *AuthRepository) FindSessionsByUserID(userID string) ([]models.Session, error) {
	var sessions []models.Session
	err := r.db.
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *AuthRepository) DeleteSessionByID(tx *gorm.DB, id string) error {
	return tx.Where("id = ?", id).Delete(&models.Session{}).Error
}

func (r *AuthRepository) DeleteAllSessionsByUserID(tx *gorm.DB, userID string) error {
	return tx.Where("user_id = ?", userID).Delete(&models.Session{}).Error
}


func (r *AuthRepository) UpsertOAuthAccount(tx *gorm.DB, account *models.ProviderAccount) error {
	return tx.
		Where(models.ProviderAccount{
			UserID:            account.UserID,
			Provider:          account.Provider,
			ProviderAccountID: account.ProviderAccountID,
		}).
		Assign(models.ProviderAccount{
			AccessToken:  account.AccessToken,
			RefreshToken: account.RefreshToken,
			ExpiresAt:    account.ExpiresAt,
		}).
		FirstOrCreate(account).Error
}

func (r *AuthRepository) CreateOAuthUser(tx *gorm.DB, user *models.User) error {
	return tx.Create(user).Error
}

func (r *AuthRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}