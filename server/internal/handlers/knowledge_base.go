package handlers

import (
	"log"

	"server/internal/config"
	"server/internal/config/constant"
	"server/internal/dto"
	"server/internal/middleware"
	"server/internal/services"
	"server/pkg/pagination"
	"server/pkg/response"
	"server/pkg/validator"

	"github.com/gin-gonic/gin"
)

type KnowledgeBaseHandler struct {
	svc *services.KnowledgeBaseService
	cfg *config.Config
}

func NewKnowledgeBaseHandler(svc *services.KnowledgeBaseService, cfg *config.Config) *KnowledgeBaseHandler {
	return &KnowledgeBaseHandler{svc: svc, cfg: cfg}
}

func (h *KnowledgeBaseHandler) Create(c *gin.Context) {
	var req dto.CreateKnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constant.ErrBadRequest, constant.CodeBadRequest, validator.ExtractErrors(err))
		return
	}

	user := middleware.GetAuthUser(c)
	log.Printf("[kb handler] Create request — user: %s", user.UserID)

	result, err := h.svc.Create(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[kb handler] Create success — id: %s", result.ID)
	response.Created(c, constant.CreateKnowledgeBaseSuccess, result)
}

func (h *KnowledgeBaseHandler) List(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	log.Printf("[kb handler] List request — user: %s", user.UserID)

	params := pagination.ParseQuery(c)
	items, total, err := h.svc.List(params)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	meta := response.PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		TotalItems: int(total),
		TotalPages: (int(total) + params.Limit - 1) / params.Limit,
	}

	response.OK(c, constant.ListKnowledgeBaseSuccess, items, meta)
}

func (h *KnowledgeBaseHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	user := middleware.GetAuthUser(c)
	log.Printf("[kb handler] GetByID request — id: %s, user: %s", id, user.UserID)

	result, err := h.svc.GetByID(id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[kb handler] GetByID success — id: %s", id)
	response.OK(c, constant.GetKnowledgeBaseSuccess, result)
}

func (h *KnowledgeBaseHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateKnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constant.ErrBadRequest, constant.CodeBadRequest, validator.ExtractErrors(err))
		return
	}

	user := middleware.GetAuthUser(c)
	log.Printf("[kb handler] Update request — id: %s, user: %s", id, user.UserID)

	result, err := h.svc.Update(id, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[kb handler] Update success — id: %s", id)
	response.OK(c, constant.UpdateKnowledgeBaseSuccess, result)
}

func (h *KnowledgeBaseHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	user := middleware.GetAuthUser(c)
	log.Printf("[kb handler] Delete request — id: %s, user: %s", id, user.UserID)

	if err := h.svc.Delete(id); err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[kb handler] Delete success — id: %s", id)
	response.OK(c, constant.DeleteKnowledgeBaseSuccess)
}
