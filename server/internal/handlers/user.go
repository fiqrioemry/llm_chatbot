package handlers

import (
	"server/internal/config"
	"server/internal/config/constant"
	"server/internal/dto"
	"server/internal/middleware"
	"server/internal/services"
	"server/pkg/response"
	"server/pkg/validator"

	"github.com/gin-gonic/gin"
)
type UserHandler struct {
	svc *services.UserService
	cfg *config.Config
}


func NewUserHandler(svc *services.UserService, cfg *config.Config) *UserHandler {
	return &UserHandler{svc: svc, cfg: cfg}
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c,
			constant.ErrBadRequest,
			constant.CodeBadRequest,
			validator.ExtractErrors(err),
		)
		return
	}

	if err := h.svc.UpdateProfile(c, user.UserID, req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, constant.UpdateProfileSuccess)
}

func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	user := middleware.GetAuthUser(c)

	file, err := c.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, constant.ErrBadRequest, constant.CodeBadRequest, "Avatar file is required")
		return
	}

	if err := h.svc.UpdateAvatar(c, user.UserID, file); err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, constant.UpdateAvatarSuccess)
}

// func (h *Handler) AddAddress(c *gin.Context) {
// 	user := middleware.GetAuthUser(c)
// 	var req AddAddressRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		response.BadRequest(c,
// 			constant.ErrBadRequest,
// 			constant.CodeBadRequest,
// 			validator.ExtractErrors(err),
// 		)
// 		return
// 	}

// 	if err := h.svc.AddAddress(c, user.UserID, req); err != nil {
// 		response.HandleError(c, err)
// 		return
// 	}

// 	response.OK(c, constant.AddAddressSuccess)
// }