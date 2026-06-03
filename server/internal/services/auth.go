package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"server/internal/config"
	"server/internal/config/constant"
	"server/internal/dto"
	"server/internal/lib"
	"server/internal/models"
	"server/internal/repositories"
	"server/pkg/cache"
	"server/pkg/crypto"
	"server/pkg/jwt"
	"server/pkg/response"
	"server/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

 
 
 
type AuthService struct {
	authRepo   repositories.AuthRepositoryContract
	userRepo repositories.UserRepositoryContract
	cfg    *config.Config
	cache  *cache.Client
	mailer *lib.Mailer        
	oauth  *lib.OAuthRegistry  
	db	 *gorm.DB
}

func NewAuthService(
	repo repositories.AuthRepositoryContract,
	userRepo repositories.UserRepositoryContract,
	cfg *config.Config,
	cache *cache.Client,
	mailer *lib.Mailer,
	oauth *lib.OAuthRegistry,
	db *gorm.DB,
) *AuthService {
	return &AuthService{authRepo: repo, userRepo :userRepo, cfg: cfg, cache: cache, mailer: mailer, oauth: oauth, db: db}
}

 
func (s *AuthService) Register(c *gin.Context, req dto.RegisterRequest) (string, error) {
	existing, err := s.authRepo.FindUserByEmail(req.Email)
	if err != nil {
		
		return "", response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}
	if existing != nil {
		return "", response.BadRequestErr(constant.ErrEmailExists, constant.CodeEmailExists)
	}

	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		return "", response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}

	username    := utils.GenerateUsername(req.Fullname)
	userID      := utils.NewUUID()
	rawToken, err := crypto.GenerateOpaqueToken()
	if err != nil {
		return "", response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}
	hashedToken := crypto.HashToken(rawToken)

	err = s.authRepo.Transaction(func(tx *gorm.DB) error {
		avatarURL := utils.GenerateAvatarURL(username)
		user := &models.User{
			ID:           userID,
			Email:        req.Email,
			Fullname:     req.Fullname,
			PasswordHash: &hashedPassword,
			Avatar:       &avatarURL,
			Status:       models.UserStatusInactive,
			Role:         models.UserRoleUser,
		}
		if err := s.authRepo.CreateUser(tx, user); err != nil {
			return err
		}


		v := &models.UserVerification{
			ID:        utils.NewUUID(),
			UserID:    userID,
			Token:     hashedToken,
			Type:      "EMAIL_VERIFICATION",
			ExpiresAt: time.Now().Add(time.Duration(s.cfg.Auth.VerifyEmailExpSec) * time.Second),
		}
		if err := s.authRepo.CreateVerification(tx, v); err != nil {
			return err
		}

		metadata := map[string]string{"email": req.Email, "user_agent": c.Request.UserAgent(), "ip_address": c.ClientIP()}
		metadataBytes, _ := json.Marshal(metadata)

		return s.userRepo.CreateActivityLog(tx, &models.UserActivityLog{
			ID:     utils.NewUUID(),
			UserID: userID,
			Action: constant.ActionRegister,
			Metadata:  string(metadataBytes),
		})
	})
	if err != nil {
		return "", response.InternalErr(constant.ErrUserCreationFailed, constant.CodeUserCreationFailed)
	}

	go func() {
		link := fmt.Sprintf("%s/verify-email?token=%s", s.cfg.App.ClientURL, rawToken)
		s.mailer.SendVerificationLink(req.Email, s.cfg.Mail.EmailVerificationSubject, link)
	}()

	return req.Email, nil
}


func (s *AuthService) VerifyEmail(token string) error {
	if token == "" {
		return response.BadRequestErr(constant.ErrTokenRequired, constant.CodeTokenRequired)
	}

	hashed := crypto.HashToken(token)
	v, err := s.authRepo.FindVerification(hashed, "EMAIL_VERIFICATION")
	if err != nil {
		return response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}
	if v == nil {
		return response.BadRequestErr(constant.ErrInvalidToken, constant.CodeInvalidToken)
	}

	return s.authRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.authRepo.MarkVerificationUsed(tx, v.ID); err != nil {
			return err
		}
		return s.authRepo.UpdateUserStatus(tx, v.UserID, models.UserStatusActive)
	})
}

 
func (s *AuthService) ResendVerificationEmail(email string) error {
	user, err := s.authRepo.FindUserByEmail(email)
	if err != nil {
		return response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}
	// return 200 meski email tidak ditemukan — prevent email enumeration
	if user == nil {
		return nil
	}
	if user.VerifiedAt != nil {
		return response.BadRequestErr(constant.ErrEmailAlreadyVerified, constant.CodeEmailAlreadyVerified)
	}

	rawToken, _ := crypto.GenerateOpaqueToken()
	hashedToken := crypto.HashToken(rawToken)

	v := &models.UserVerification{
		ID:        utils.NewUUID(),
		UserID:    user.ID,
		Token:     hashedToken,
		Type:      "EMAIL_VERIFICATION",
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.Auth.VerifyEmailExpSec) * time.Second),
	}
	s.authRepo.CreateVerification(s.db, v)

	go func() {
		link := fmt.Sprintf("%s/verify-email?token=%s", s.cfg.App.ClientURL, rawToken)
		s.mailer.SendVerificationLink(email, s.cfg.Mail.EmailVerificationSubject, link)
	}()

	return nil
}

 
func (s *AuthService) Login(c *gin.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.authRepo.FindUserByEmail(req.Email)
	if err != nil {
		return nil, response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}
	if user == nil || user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, response.BadRequestErr(constant.ErrInvalidCredentials, constant.CodeInvalidCredentials)
	}
	if user.Status == models.UserStatusInactive || user.VerifiedAt == nil {
		return nil, response.BadRequestErr(constant.ErrEmailNotVerified, constant.CodeEmailNotVerified)
	}
	if user.Status == models.UserStatusBanned {
		return nil, response.ForbiddenErr(constant.ErrAccountBanned, constant.CodeAccountBanned)
	}
	if !crypto.VerifyPassword(req.Password, *user.PasswordHash) {
		return nil, response.BadRequestErr(constant.ErrInvalidCredentials, constant.CodeInvalidCredentials)
	}

	return s.createSession(c, user)
}

 
func (s *AuthService) Logout(sessionID string) error {
	return s.authRepo.Transaction(func(tx *gorm.DB) error {
		return s.authRepo.DeleteSessionByID(tx, sessionID)
	})
}

func (s *AuthService) LogoutAll(userID string) error {
	return s.authRepo.Transaction(func(tx *gorm.DB) error {
		return s.authRepo.DeleteAllSessionsByUserID(tx, userID)
	})
}

 
func (s *AuthService) GetSessions(userID string) ([]dto.SessionResponse, error) {
	sessions, err := s.authRepo.FindSessionsByUserID(userID)
	if err != nil {
		return nil, response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}

	result := make([]dto.SessionResponse, len(sessions))
	for i, sess := range sessions {
		result[i] = dto.SessionResponse{
			ID:        sess.ID,
			UserAgent: sess.UserAgent,
			IPAddress: sess.IPAddress,
			ExpiresAt: sess.ExpiresAt,
			CreatedAt: sess.CreatedAt,
		}
	}
	return result, nil
}

func (s *AuthService) DeleteSession(sessionID string) error {
	sess, err := s.authRepo.FindSessionByID(sessionID)
	if err != nil {
		return response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}
	if sess == nil {
		return response.NotFoundErr(constant.ErrSessionNotFound, constant.CodeSessionNotFound)
	}
	return s.authRepo.DeleteSessionByID(s.db, sessionID)
}

 
func (s *AuthService) RefreshToken(c *gin.Context, rawRefreshToken string) (*dto.LoginResponse, error) {
	hashed := crypto.HashToken(rawRefreshToken)
	sess, err := s.authRepo.FindSessionByRefreshToken(hashed)
	if err != nil {
		return nil, response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}
	if sess == nil {
		return nil, response.UnauthorizedErr(constant.ErrRefreshTokenInvalid, constant.CodeRefreshTokenInvalid)
	}
	if sess.User.Status == models.UserStatusBanned {
		return nil, response.ForbiddenErr(constant.ErrAccountBanned, constant.CodeAccountBanned)
	}

	// Hapus session lama — refresh token rotation, satu RT hanya valid sekali pakai
	s.authRepo.DeleteSessionByID(s.db, sess.ID)

	return s.createSession(c, &sess.User)
}

func (s *AuthService) ChangePassword(userID string, req dto.ChangePasswordRequest) error {
	hash, err := s.authRepo.GetPasswordHash(userID)
	if err != nil || hash == "" {
		return response.NotFoundErr(constant.ErrUserNotFound, constant.CodeUserNotFound)
	}
	if !crypto.VerifyPassword(req.CurrentPassword, hash) {
		return response.BadRequestErr(constant.ErrInvalidCredentials, constant.CodeInvalidCredentials)
	}

	newHash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		return response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}

	return s.authRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.authRepo.UpdatePassword(tx, userID, newHash); err != nil {
			return err
		}
		return s.authRepo.DeleteAllSessionsByUserID(tx, userID)
	})
}

 
func (s *AuthService) ForgotPassword(email string) error {
	user, err := s.authRepo.FindUserByEmail(email)
	if err != nil {
		return response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}
	if user == nil {
		return nil // prevent email enumeration
	}

	rawToken, _ := crypto.GenerateOpaqueToken()
	hashedToken := crypto.HashToken(rawToken)

	v := &models.UserVerification{
		ID:        utils.NewUUID(),
		UserID:    user.ID,
		Token:     hashedToken,
		Type:      "PASSWORD_RESET",
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.Auth.ResetPasswordExpSec) * time.Second),
	}
	s.authRepo.CreateVerification(s.db, v)

	go func() {
		link := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.App.ClientURL, rawToken)
		s.mailer.SendVerificationLink(email, s.cfg.Mail.PasswordResetSubject, link)
	}()

	return nil
}

 
func (s *AuthService) ResetPassword(token string, req dto.ResetPasswordRequest) error {
	hashed := crypto.HashToken(token)
	v, err := s.authRepo.FindVerification(hashed, "PASSWORD_RESET")
	if err != nil {
		return response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}
	if v == nil {
		return response.BadRequestErr(constant.ErrInvalidToken, constant.CodeInvalidToken)
	}

	newHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}

	return s.authRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.authRepo.MarkVerificationUsed(tx, v.ID); err != nil {
			return err
		}
		if err := s.authRepo.UpdatePassword(tx, v.UserID, newHash); err != nil {
			return err
		}
		return s.authRepo.DeleteAllSessionsByUserID(tx, v.UserID)
	})
}
 
func (s *AuthService) GetMe(userID string) (*dto.MeResponse, error) {
	user, err := s.authRepo.FindUserByID(userID)
	if err != nil || user == nil {
		return nil, response.NotFoundErr(constant.ErrUserNotFound, constant.CodeUserNotFound)
	}

	return &dto.MeResponse{
		ID:                   user.ID,
		Fullname:             user.Fullname,
		Email:                user.Email,
		Avatar:               user.Avatar,
		Role:                 string(user.Role),
		LastLogin:            user.LastLogin,
		JoinedAt:             user.CreatedAt,
		LastChangePasswordAt: user.LastChangePasswordAt,
	}, nil
}

 
func (s *AuthService) GetOAuthURL(provider string) (string, error) {
	p, err := s.oauth.Get(provider)
	if err != nil {
		return "", response.BadRequestErr(err.Error(), constant.CodeBadRequest)
	}

	state, _ := crypto.GenerateOpaqueToken()
	key := fmt.Sprintf("oauth:state:%s:%s", provider, state)
	s.cache.Set(context.Background(), key, "1", s.cfg.Auth.OAuthStateTTLSec)

	return p.GetAuthURL(state), nil
}

func (s *AuthService) HandleOAuthCallback(c *gin.Context, provider, code, state string) (*dto.LoginResponse, error) {
	// 1. Validasi CSRF state
	stateKey := fmt.Sprintf("oauth:state:%s:%s", provider, state)
	exists, err := s.cache.Exists(context.Background(), stateKey)
	if err != nil || !exists {
		return nil, response.BadRequestErr(constant.ErrOAuthStateMismatch, constant.CodeOAuthStateMismatch)
	}
	s.cache.Del(context.Background(), stateKey)

	// 2. Exchange code → tokens
	p, err := s.oauth.Get(provider)
	if err != nil {
		return nil, response.BadRequestErr(err.Error(), constant.CodeBadRequest)
	}

	tokens, err := p.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		return nil, response.BadRequestErr(constant.ErrOAuthFailed, constant.CodeOAuthFailed)
	}

	oauthUser, err := p.GetUserInfo(c.Request.Context(), tokens.AccessToken)
	if err != nil {
		return nil, response.BadRequestErr(constant.ErrOAuthFailed, constant.CodeOAuthFailed)
	}
	if oauthUser.Email == "" {
		return nil, response.BadRequestErr(constant.ErrOAuthEmailNotProvided, constant.CodeOAuthEmailNotProvided)
	}

	// 3. Resolve user
	oauthProvider := models.OAuthProvider(provider)
	user, err := s.authRepo.FindUserByOAuthProvider(oauthProvider, oauthUser.ProviderAccountID)
	if err != nil {
		return nil, response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}

	if user == nil {
		existing, _ := s.authRepo.FindUserByEmail(oauthUser.Email)
		if existing != nil {
			user = existing
		} else {
			username  := utils.GenerateUsername(oauthUser.Name)
			avatarURL := oauthUser.Avatar
			if avatarURL == "" {
				avatarURL = utils.GenerateAvatarURL(username)
			}
			now     := time.Now()
			newUser := &models.User{
				ID:         utils.NewUUID(),
				Email:      oauthUser.Email,
				Fullname:   oauthUser.Name,
				Avatar:     &avatarURL,
				Status:     models.UserStatusActive,
				Role:       models.UserRoleUser,
				VerifiedAt: &now,
			}
			s.authRepo.Transaction(func(tx *gorm.DB) error {
				return s.authRepo.CreateOAuthUser(tx, newUser)
			})
			user = newUser
		}
	}

	if user.Status == models.UserStatusBanned {
		return nil, response.ForbiddenErr(constant.ErrAccountBanned, constant.CodeAccountBanned)
	}

	// 4. Upsert OAuth account record
	s.authRepo.UpsertOAuthAccount(s.db, &models.ProviderAccount{
		ID:                utils.NewUUID(),
		UserID:            user.ID,
		Provider:          oauthProvider,
		ProviderAccountID: oauthUser.ProviderAccountID,
		AccessToken:       tokens.AccessToken,
		RefreshToken:      tokens.RefreshToken,
		ExpiresAt:         tokens.ExpiresAt,
	})

	// 5. Buat session baru
	return s.createSession(c, user)
}

 
func (s *AuthService) createSession(c *gin.Context, user *models.User) (*dto.LoginResponse, error) {
	rawRT, err := crypto.GenerateOpaqueToken()
	if err != nil {
		return nil, response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}
	hashedRT := crypto.HashToken(rawRT)

	rtExpiry := time.Now().Add(time.Duration(s.cfg.Auth.RefreshTokenExpSec) * time.Second)
	atExpiry := time.Now().Add(time.Duration(s.cfg.Auth.AccessTokenExpSec) * time.Second)

	userAgent, ipAddress := "", ""
	if c != nil {
		userAgent = c.Request.UserAgent()
		ipAddress = c.ClientIP()
	}

	session := &models.Session{
		ID:           utils.NewUUID(),
		UserID:       user.ID,
		RefreshToken: hashedRT,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    rtExpiry,
	}

	err = s.authRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.authRepo.CreateSession(tx, session); err != nil {
			return err
		}
		if err := s.authRepo.UpdateLastLogin(tx, user.ID); err != nil {
			return err
		}
		metadata := map[string]string{"ip_address": ipAddress, "user_agent": userAgent}
		metadataBytes, _ := json.Marshal(metadata)
		return s.userRepo.CreateActivityLog(tx, &models.UserActivityLog{
			ID:     utils.NewUUID(),
			UserID: user.ID,
			Action: constant.ActionLogin,
			Metadata:  string(metadataBytes),
		})
	})
	if err != nil {
		return nil, response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}

 
	accessToken, err := jwt.GenerateAccessToken(
		s.cfg.Auth.AccessTokenSecret,
		s.cfg.Auth.AccessTokenExpSec,
		user.ID,
		user.Fullname,
		string(user.Role),
	)
	if err != nil {
		return nil, response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}

	return &dto.LoginResponse{
		Tokens: dto.TokenPair{
			AccessToken:           accessToken,
			RefreshToken:          rawRT,  
			AccessTokenExpiresAt:  atExpiry,
			RefreshTokenExpiresAt: rtExpiry,
		},
		User: dto.UserSummary{
			ID:       user.ID,
			Fullname: user.Fullname,
			Email:    user.Email,
			Avatar:   user.Avatar,
		},
	}, nil
}