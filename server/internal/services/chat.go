package services

import (
	"context"
	"log"
	"strings"
	"time"

	"server/internal/config"
	"server/internal/config/constant"
	"server/internal/dto"
	"server/internal/models"
	"server/internal/repositories"
	"server/pkg/pagination"
	"server/pkg/response"
	"server/providers/openai"
)

type ChatService struct {
	chatRepo repositories.ChatRepositoryContract
	kbRepo   repositories.KnowledgeBaseRepositoryContract
	ai       *openai.Client
	cfg      *config.Config
}

func NewChatService(
	chatRepo repositories.ChatRepositoryContract,
	kbRepo repositories.KnowledgeBaseRepositoryContract,
	ai *openai.Client,
	cfg *config.Config,
) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
		kbRepo:   kbRepo,
		ai:       ai,
		cfg:      cfg,
	}
}

func (s *ChatService) CreateConversation(userID, kbID string) (*dto.ConversationResponse, error) {
	kb, err := s.kbRepo.FindByID(kbID)
	if err != nil {
		return nil, response.InternalErr(err.Error(), "KB_FETCH_FAIL")
	}
	if kb == nil {
		return nil, response.NotFoundErr(constant.ErrKnowledgeBaseNotFound, constant.CodeKnowledgeBaseNotFound)
	}

	conv := models.Conversation{
		UserID:          userID,
		KnowledgeBaseID: kbID,
		Title:           "New Conversation",
	}

	if err := s.chatRepo.CreateConversation(&conv); err != nil {
		return nil, response.InternalErr(err.Error(), "CONV_CREATE_FAIL")
	}

	log.Printf("[chat service] CreateConversation — user: %s, kb: %s", userID, kbID)

	res := s.convToResponse(conv)
	return &res, nil
}

func (s *ChatService) ListConversations(userID string, params pagination.PaginationParams) ([]dto.ConversationResponse, int64, error) {
	convs, total, err := s.chatRepo.FindConversations(userID, params)
	if err != nil {
		return nil, 0, response.InternalErr(err.Error(), "CONV_LIST_FAIL")
	}

	result := make([]dto.ConversationResponse, 0, len(convs))
	for _, c := range convs {
		result = append(result, s.convToResponse(c))
	}

	return result, total, nil
}

func (s *ChatService) DeleteConversation(userID, convID string) error {
	conv, err := s.chatRepo.FindConversationByID(convID)
	if err != nil {
		return response.InternalErr(err.Error(), "CONV_FETCH_FAIL")
	}
	if conv == nil {
		return response.NotFoundErr(constant.ErrConversationNotFound, constant.CodeConversationNotFound)
	}
	if conv.UserID != userID {
		return response.ForbiddenErr(constant.ErrConversationForbidden, constant.CodeConversationForbidden)
	}

	if err := s.chatRepo.DeleteConversation(convID); err != nil {
		return response.InternalErr(err.Error(), "CONV_DELETE_FAIL")
	}

	return nil
}

func (s *ChatService) SendMessage(ctx context.Context, userID, convID, content string, onChunk func(string)) (*dto.MessageResponse, error) {
	// 1. Find conversation, verify ownership
	conv, err := s.chatRepo.FindConversationByID(convID)
	if err != nil {
		return nil, response.InternalErr(err.Error(), "CONV_FETCH_FAIL")
	}
	if conv == nil {
		return nil, response.NotFoundErr(constant.ErrConversationNotFound, constant.CodeConversationNotFound)
	}
	if conv.UserID != userID {
		return nil, response.ForbiddenErr(constant.ErrConversationForbidden, constant.CodeConversationForbidden)
	}

	// 2. Load KB
	kb, err := s.kbRepo.FindByID(conv.KnowledgeBaseID)
	if err != nil || kb == nil {
		return nil, response.InternalErr("knowledge base not found", "KB_FETCH_FAIL")
	}

	// 3. Save user message
	userMsg := models.Message{
		ConversationID: convID,
		Role:           models.MessageRoleUser,
		Content:        content,
	}
	if err := s.chatRepo.CreateMessage(&userMsg); err != nil {
		return nil, response.InternalErr(err.Error(), "MSG_CREATE_FAIL")
	}

	// 4. Embed query
	embedding, err := s.ai.Embed(ctx, content)
	if err != nil {
		return nil, err
	}

	// 5. Search similar chunks (topK=5)
	similarChunks, err := s.chatRepo.SearchSimilarChunks(conv.KnowledgeBaseID, embedding, 5)
	if err != nil {
		return nil, response.InternalErr(err.Error(), "VECTOR_SEARCH_FAIL")
	}

	// 6. Log vector search result
	topScore := float32(0)
	if len(similarChunks) > 0 {
		topScore = similarChunks[0].SimilarityScore
	}
	log.Printf("[chat service] VectorSearch — found %d chunks, top score: %.3f", len(similarChunks), topScore)

	// 7. Find recent messages for history
	recentMsgs, err := s.chatRepo.FindRecentMessages(convID, 10)
	if err != nil {
		return nil, response.InternalErr(err.Error(), "MSG_FETCH_FAIL")
	}

	// 8. Build context string from chunks
	var contextParts []string
	for _, chunk := range similarChunks {
		contextParts = append(contextParts, chunk.Content)
	}
	contextStr := strings.Join(contextParts, "\n\n")

	// 9. Build messages
	systemPrompt := "You are a helpful assistant. Answer based ONLY on the provided context. If the answer is not in the context, say you don't know."
	historyMsgs := mapHistoryToOpenAI(recentMsgs)
	msgs := openai.BuildRAGMessages(systemPrompt, contextStr, content, historyMsgs)

	// 10. Record start time
	startTime := time.Now()

	// 11. Collect full response using ChatStream + onChunk callback
	var fullContent strings.Builder
	err = s.ai.ChatStream(ctx, msgs, nil, func(delta string) {
		if delta != "" {
			fullContent.WriteString(delta)
			onChunk(delta)
		}
	})
	if err != nil {
		return nil, err
	}

	// 12. Compute latencyMs, estimate tokens
	latencyMs := int(time.Since(startTime).Milliseconds())
	tokensUsed := len(fullContent.String()) / 4

	// 13. Log LLM call
	log.Printf("[chat service] LLMCall — tokens: %d, latency: %dms", tokensUsed, latencyMs)

	// 14. Save assistant message
	assistantMsg := models.Message{
		ConversationID: convID,
		Role:           models.MessageRoleAssistant,
		Content:        fullContent.String(),
		TokensUsed:     tokensUsed,
		LatencyMs:      latencyMs,
	}
	if err := s.chatRepo.CreateMessage(&assistantMsg); err != nil {
		return nil, response.InternalErr(err.Error(), "MSG_CREATE_FAIL")
	}

	// 15. Save sources from chunks
	sources := make([]models.MessageSource, 0, len(similarChunks))
	for _, chunk := range similarChunks {
		sources = append(sources, models.MessageSource{
			MessageID:       assistantMsg.ID,
			ChunkID:         chunk.ChunkID,
			SimilarityScore: chunk.SimilarityScore,
			Rank:            chunk.Rank,
		})
	}
	if err := s.chatRepo.CreateMessageSources(sources); err != nil {
		log.Printf("[chat service] failed to save message sources: %v", err)
	}

	// 16. Auto-title if first message (user message was the first saved, so count = 2 now)
	if conv.Title == "New Conversation" {
		title := content
		if len(title) > 60 {
			title = title[:60]
		}
		_ = s.chatRepo.UpdateConversationTitle(convID, title)
	}

	// 17. Log completion
	log.Printf("[chat service] SendMessage — completed: conversation: %s", convID)

	// 18. Return MessageResponse with sources populated
	sourceResponses := make([]dto.SourceResponse, 0, len(similarChunks))
	for _, chunk := range similarChunks {
		sourceResponses = append(sourceResponses, dto.SourceResponse{
			ChunkID:          chunk.ChunkID,
			Content:          chunk.Content,
			SimilarityScore:  chunk.SimilarityScore,
			Rank:             chunk.Rank,
			DocumentFilename: chunk.DocumentFilename,
		})
	}

	return &dto.MessageResponse{
		ID:             assistantMsg.ID,
		ConversationID: convID,
		Role:           models.MessageRoleAssistant,
		Content:        fullContent.String(),
		TokensUsed:     tokensUsed,
		LatencyMs:      latencyMs,
		Sources:        sourceResponses,
		CreatedAt:      assistantMsg.CreatedAt,
	}, nil
}

func (s *ChatService) ListMessages(userID, convID string, params pagination.PaginationParams) ([]dto.MessageResponse, int64, error) {
	// Verify conv belongs to user
	conv, err := s.chatRepo.FindConversationByID(convID)
	if err != nil {
		return nil, 0, response.InternalErr(err.Error(), "CONV_FETCH_FAIL")
	}
	if conv == nil {
		return nil, 0, response.NotFoundErr(constant.ErrConversationNotFound, constant.CodeConversationNotFound)
	}
	if conv.UserID != userID {
		return nil, 0, response.ForbiddenErr(constant.ErrConversationForbidden, constant.CodeConversationForbidden)
	}

	msgs, total, err := s.chatRepo.FindMessages(convID, params)
	if err != nil {
		return nil, 0, response.InternalErr(err.Error(), "MSG_LIST_FAIL")
	}

	result := make([]dto.MessageResponse, 0, len(msgs))
	for _, msg := range msgs {
		msgSources, err := s.chatRepo.FindMessageSources(msg.ID)
		if err != nil {
			msgSources = nil
		}

		sourceResponses := make([]dto.SourceResponse, 0, len(msgSources))
		for _, src := range msgSources {
			sourceResponses = append(sourceResponses, dto.SourceResponse{
				ChunkID:          src.ChunkID,
				Content:          src.Chunk.Content,
				SimilarityScore:  src.SimilarityScore,
				Rank:             src.Rank,
				DocumentFilename: src.Chunk.Document.Filename,
			})
		}

		result = append(result, dto.MessageResponse{
			ID:             msg.ID,
			ConversationID: msg.ConversationID,
			Role:           msg.Role,
			Content:        msg.Content,
			TokensUsed:     msg.TokensUsed,
			LatencyMs:      msg.LatencyMs,
			Sources:        sourceResponses,
			CreatedAt:      msg.CreatedAt,
		})
	}

	return result, total, nil
}

func (s *ChatService) convToResponse(c models.Conversation) dto.ConversationResponse {
	return dto.ConversationResponse{
		ID:              c.ID,
		UserID:          c.UserID,
		KnowledgeBaseID: c.KnowledgeBaseID,
		Title:           c.Title,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}

func mapHistoryToOpenAI(msgs []models.Message) []openai.ChatMessage {
	result := make([]openai.ChatMessage, 0, len(msgs))
	for _, m := range msgs {
		var role openai.Role
		switch m.Role {
		case models.MessageRoleUser:
			role = openai.RoleUser
		case models.MessageRoleAssistant:
			role = openai.RoleAssistant
		default:
			role = openai.RoleUser
		}
		result = append(result, openai.ChatMessage{
			Role:    role,
			Content: m.Content,
		})
	}
	return result
}
