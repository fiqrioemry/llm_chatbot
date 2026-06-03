package services

import (
	"context"
	"encoding/json"
	"log"
	"mime/multipart"

	"server/internal/config/constant"
	"server/internal/dto"
	"server/internal/models"
	"server/internal/repositories"
	"server/pkg/cache"
	"server/pkg/response"
	"server/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo    *repositories.UserRepository
	fileSvc *FileService
	cache   *cache.Client
}

func NewUserService(userRepo *repositories.UserRepository, fileSvc *FileService, cache *cache.Client) *UserService {
	return &UserService{
		userRepo: userRepo,
		fileSvc: fileSvc,
		cache:   cache,	
	}
}

func (s *UserService) UpdateProfile(c *gin.Context, userID string, req dto.UpdateProfileRequest) error {
	err, user := s.userRepo.GetByID(userID)
	if err != nil {
		return response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}
	if user == nil {
		return response.NotFoundErr(constant.ErrUserNotFound, constant.CodeUserNotFound)
	}
	if req.Fulname != "" {
		user.Fullname = req.Fulname
	}


	if err := s.userRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.UpdateProfile(tx, userID, user); err != nil {
			return response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
		}
		

		metadata := map[string]interface{}{
			"previous_data": map[string]interface{}{
				"fullname": user.Fullname,
			},
			"current_data": map[string]interface{}{
				"fullname": req.Fulname,
			},
			"ip_address": c.ClientIP(),
			"user_agent": c.Request.UserAgent(),			
		}
		
		metadataBytes, _ := json.Marshal(metadata)
		return s.userRepo.CreateActivityLog(tx, &models.UserActivityLog{
			ID:     utils.NewUUID(),
			UserID: userID,
			Action: constant.ActionUpdateProfile,
			Metadata: metadataBytes,
		})
	}); err != nil {
		return err
	}

	s.cache.Del(c, "user:"+userID)
	return nil
}

func (s *UserService) UpdateAvatar(c *gin.Context, userID string, fileHeader *multipart.FileHeader) error {
    err, user := s.userRepo.GetByID(userID)
    if err != nil {
        return response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
    }
    if user == nil {
        return response.NotFoundErr(constant.ErrUserNotFound, constant.CodeUserNotFound)
    }

    createdBy := userID
    fileData, err := s.fileSvc.GenerateFileRecord(fileHeader, "avatars", &createdBy)
    if err != nil {
        return response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
    }
    fileData.IsUsed = true
    fileData.TargetID = &userID


    if len(fileData.FileBuffer) == 0 {
        return response.BadRequestErr("file is empty", constant.CodeBadRequest)
    }

    if err := s.userRepo.Transaction(func(tx *gorm.DB) error {
        saved, err := s.fileSvc.SaveRecord(tx, fileData)
        if err != nil {
            return err
        }

        if err := s.userRepo.UpdateAvatar(tx, userID, saved.URL); err != nil {
            return err
        }

        metadata := map[string]interface{}{
            "ip_address": c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
            "file_id":    saved.ID,
        }
        metaBytes, _ := json.Marshal(metadata)
        return s.userRepo.CreateActivityLog(tx, &models.UserActivityLog{
            ID:       utils.NewUUID(),
            UserID:   userID,
            Action:   constant.ActionUpdateAvatar,
            Metadata: metaBytes,
        })
    }); err != nil {    

        return err
    }


	if err := s.fileSvc.UploadToStorage(context.Background(), fileData); err != nil {
		log.Printf("[UpdateAvatar] upload failed: %v", err)
		return response.InternalErr(constant.ErrInternalServerError, constant.CodeInternalServerError)
	}

    return nil
}
