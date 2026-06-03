package handlers

import (
	"log"

	"server/internal/config"
	"server/internal/config/constant"
	"server/internal/middleware"
	"server/internal/services"
	"server/pkg/pagination"
	"server/pkg/response"

	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	svc *services.DocumentService
	cfg *config.Config
}

func NewDocumentHandler(svc *services.DocumentService, cfg *config.Config) *DocumentHandler {
	return &DocumentHandler{svc: svc, cfg: cfg}
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	kbID := c.Param("kbId")
	user := middleware.GetAuthUser(c)
	log.Printf("[document handler] Upload request — kb: %s, user: %s", kbID, user.UserID)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File is required", "FILE_REQUIRED")
		return
	}

	result, svcErr := h.svc.Upload(c.Request.Context(), kbID, fileHeader)
	if svcErr != nil {
		response.HandleError(c, svcErr)
		return
	}

	log.Printf("[document handler] Upload success — doc: %s, kb: %s", result.ID, kbID)
	response.Created(c, constant.UploadDocumentSuccess, result)
}

func (h *DocumentHandler) List(c *gin.Context) {
	kbID := c.Param("kbId")
	user := middleware.GetAuthUser(c)
	log.Printf("[document handler] List request — kb: %s, user: %s", kbID, user.UserID)

	status := c.Query("status")
	params := pagination.ParseQuery(c)

	items, total, err := h.svc.List(kbID, status, params)
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

	response.OK(c, constant.ListDocumentsSuccess, items, meta)
}

func (h *DocumentHandler) GetByID(c *gin.Context) {
	kbID := c.Param("kbId")
	docID := c.Param("id")
	user := middleware.GetAuthUser(c)
	log.Printf("[document handler] GetByID request — doc: %s, kb: %s, user: %s", docID, kbID, user.UserID)

	result, err := h.svc.GetByID(kbID, docID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[document handler] GetByID success — doc: %s", docID)
	response.OK(c, constant.GetDocumentSuccess, result)
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	kbID := c.Param("kbId")
	docID := c.Param("id")
	user := middleware.GetAuthUser(c)
	log.Printf("[document handler] Delete request — doc: %s, kb: %s, user: %s", docID, kbID, user.UserID)

	if err := h.svc.Delete(c.Request.Context(), kbID, docID); err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[document handler] Delete success — doc: %s", docID)
	response.OK(c, constant.DeleteDocumentSuccess)
}
