package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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

type ChatHandler struct {
	svc *services.ChatService
	cfg *config.Config
}

func NewChatHandler(svc *services.ChatService, cfg *config.Config) *ChatHandler {
	return &ChatHandler{svc: svc, cfg: cfg}
}

func (h *ChatHandler) CreateConversation(c *gin.Context) {
	var req dto.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constant.ErrBadRequest, constant.CodeBadRequest, validator.ExtractErrors(err))
		return
	}

	user := middleware.GetAuthUser(c)
	log.Printf("[chat handler] CreateConversation request — user: %s, kb: %s", user.UserID, req.KnowledgeBaseID)

	result, err := h.svc.CreateConversation(user.UserID, req.KnowledgeBaseID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[chat handler] CreateConversation success — id: %s", result.ID)
	response.Created(c, constant.CreateConversationSuccess, result)
}

func (h *ChatHandler) ListConversations(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	log.Printf("[chat handler] ListConversations request — user: %s", user.UserID)

	params := pagination.ParseQuery(c)
	items, total, err := h.svc.ListConversations(user.UserID, params)
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

	response.OK(c, constant.ListConversationsSuccess, items, meta)
}

func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	convID := c.Param("id")
	user := middleware.GetAuthUser(c)
	log.Printf("[chat handler] DeleteConversation request — id: %s, user: %s", convID, user.UserID)

	if err := h.svc.DeleteConversation(user.UserID, convID); err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[chat handler] DeleteConversation success — id: %s", convID)
	response.OK(c, constant.DeleteConversationSuccess)
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	convID := c.Param("id")
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constant.ErrBadRequest, constant.CodeBadRequest, validator.ExtractErrors(err))
		return
	}

	user := middleware.GetAuthUser(c)
	log.Printf("[chat handler] SendMessage request — conv: %s, user: %s", convID, user.UserID)

	// SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	result, err := h.svc.SendMessage(c.Request.Context(), user.UserID, convID, req.Content, func(chunk string) {
		fmt.Fprintf(c.Writer, "data: %s\n\n", chunk)
		c.Writer.(http.Flusher).Flush()
	})
	if err != nil {
		errData, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", errData)
		c.Writer.(http.Flusher).Flush()
		return
	}

	finalData, _ := json.Marshal(result)
	fmt.Fprintf(c.Writer, "event: done\ndata: %s\n\n", string(finalData))
	c.Writer.(http.Flusher).Flush()

	log.Printf("[chat handler] SendMessage success — conv: %s", convID)
}

func (h *ChatHandler) ListMessages(c *gin.Context) {
	convID := c.Param("id")
	user := middleware.GetAuthUser(c)
	log.Printf("[chat handler] ListMessages request — conv: %s, user: %s", convID, user.UserID)

	params := pagination.ParseQuery(c)
	items, total, err := h.svc.ListMessages(user.UserID, convID, params)
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

	response.OK(c, constant.ListMessagesSuccess, items, meta)
}
