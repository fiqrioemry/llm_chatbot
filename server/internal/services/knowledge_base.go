package services

import (
	"log"

	"server/internal/config"
	"server/internal/config/constant"
	"server/internal/dto"
	"server/internal/models"
	"server/internal/repositories"
	"server/pkg/pagination"
	"server/pkg/response"
)

type KnowledgeBaseService struct {
	repo repositories.KnowledgeBaseRepositoryContract
	cfg  *config.Config
}

func NewKnowledgeBaseService(repo repositories.KnowledgeBaseRepositoryContract, cfg *config.Config) *KnowledgeBaseService {
	return &KnowledgeBaseService{repo: repo, cfg: cfg}
}

func (s *KnowledgeBaseService) Create(req dto.CreateKnowledgeBaseRequest) (*dto.KnowledgeBaseResponse, error) {
	// Validate and apply defaults
	chunkSize := req.ChunkSize
	if chunkSize == 0 {
		chunkSize = 512
	}
	if chunkSize < 64 || chunkSize > 2048 {
		return nil, response.BadRequestErr(constant.ErrKnowledgeBaseInvalid, constant.CodeKnowledgeBaseInvalid)
	}

	chunkOverlap := req.ChunkOverlap
	if chunkOverlap == 0 {
		chunkOverlap = 64
	}
	if chunkOverlap >= chunkSize {
		return nil, response.BadRequestErr(constant.ErrKnowledgeBaseInvalid, constant.CodeKnowledgeBaseInvalid)
	}

	embedModel := req.EmbedModel
	if embedModel == "" {
		embedModel = "text-embedding-3-small"
	}

	kb := models.KnowledgeBase{
		Name:         req.Name,
		Description:  req.Description,
		EmbedModel:   embedModel,
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
	}

	if err := s.repo.Create(&kb); err != nil {
		return nil, response.InternalErr(err.Error(), "KB_CREATE_FAIL")
	}

	log.Printf("[kb service] Create — kb created: %s", kb.ID)

	res := s.toResponse(kb, 0)
	return &res, nil
}

func (s *KnowledgeBaseService) List(params pagination.PaginationParams) ([]dto.KnowledgeBaseResponse, int64, error) {
	kbs, total, err := s.repo.FindAll(params)
	if err != nil {
		return nil, 0, response.InternalErr(err.Error(), "KB_LIST_FAIL")
	}

	result := make([]dto.KnowledgeBaseResponse, 0, len(kbs))
	for _, kb := range kbs {
		count, err := s.repo.CountDocuments(kb.ID)
		if err != nil {
			count = 0
		}
		result = append(result, s.toResponse(kb, count))
	}

	return result, total, nil
}

func (s *KnowledgeBaseService) GetByID(id string) (*dto.KnowledgeBaseResponse, error) {
	kb, err := s.repo.FindByID(id)
	if err != nil {
		return nil, response.InternalErr(err.Error(), "KB_FETCH_FAIL")
	}
	if kb == nil {
		return nil, response.NotFoundErr(constant.ErrKnowledgeBaseNotFound, constant.CodeKnowledgeBaseNotFound)
	}

	count, _ := s.repo.CountDocuments(kb.ID)
	res := s.toResponse(*kb, count)
	return &res, nil
}

func (s *KnowledgeBaseService) Update(id string, req dto.UpdateKnowledgeBaseRequest) (*dto.KnowledgeBaseResponse, error) {
	kb, err := s.repo.FindByID(id)
	if err != nil {
		return nil, response.InternalErr(err.Error(), "KB_FETCH_FAIL")
	}
	if kb == nil {
		return nil, response.NotFoundErr(constant.ErrKnowledgeBaseNotFound, constant.CodeKnowledgeBaseNotFound)
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if len(updates) > 0 {
		if err := s.repo.Update(id, updates); err != nil {
			return nil, response.InternalErr(err.Error(), "KB_UPDATE_FAIL")
		}
	}

	// Re-fetch updated record
	updated, err := s.repo.FindByID(id)
	if err != nil || updated == nil {
		return nil, response.InternalErr("failed to fetch updated knowledge base", "KB_FETCH_FAIL")
	}

	log.Printf("[kb service] Update — kb updated: %s", id)

	count, _ := s.repo.CountDocuments(id)
	res := s.toResponse(*updated, count)
	return &res, nil
}

func (s *KnowledgeBaseService) Delete(id string) error {
	kb, err := s.repo.FindByID(id)
	if err != nil {
		return response.InternalErr(err.Error(), "KB_FETCH_FAIL")
	}
	if kb == nil {
		return response.NotFoundErr(constant.ErrKnowledgeBaseNotFound, constant.CodeKnowledgeBaseNotFound)
	}

	if err := s.repo.Delete(id); err != nil {
		return response.InternalErr(err.Error(), "KB_DELETE_FAIL")
	}

	log.Printf("[kb service] Delete — kb deleted: %s", id)
	return nil
}

func (s *KnowledgeBaseService) toResponse(kb models.KnowledgeBase, count int64) dto.KnowledgeBaseResponse {
	return dto.KnowledgeBaseResponse{
		ID:            kb.ID,
		Name:          kb.Name,
		Description:   kb.Description,
		EmbedModel:    kb.EmbedModel,
		ChunkSize:     kb.ChunkSize,
		ChunkOverlap:  kb.ChunkOverlap,
		DocumentCount: count,
		CreatedAt:     kb.CreatedAt,
		UpdatedAt:     kb.UpdatedAt,
	}
}
