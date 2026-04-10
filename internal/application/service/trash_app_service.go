package service

import (
	"catch/internal/application/dto"
	domainService "catch/internal/domain/service"
	"catch/internal/domain/repository"
	"fmt"
)

type TrashAppService struct {
	trashRepo    repository.TrashRepository
	configRepo   repository.ConfigRepository
	trashDomainSvc *domainService.TrashDomainService
}

func NewTrashAppService(
	trashRepo repository.TrashRepository,
	configRepo repository.ConfigRepository,
	trashDomainSvc *domainService.TrashDomainService,
) *TrashAppService {
	return &TrashAppService{
		trashRepo:    trashRepo,
		configRepo:   configRepo,
		trashDomainSvc: trashDomainSvc,
	}
}

func (s *TrashAppService) List() (*dto.TrashListResponse, error) {
	items, err := s.trashRepo.List()
	if err != nil {
		return nil, fmt.Errorf("获取回收站列表失败: %w", err)
	}

	result := make([]dto.TrashItemResponse, 0, len(items))
	for _, item := range items {
		formatted := s.trashDomainSvc.FormatTrashItem(item)
		result = append(result, dto.TrashItemResponse{
			OriginalPath:  formatted["original_path"].(string),
			TrashPath:     formatted["trash_path"].(string),
			FileName:      formatted["file_name"].(string),
			FileSize:      formatted["file_size"].(int64),
			DeletedAt:     formatted["deleted_at"].(string),
			ExpiresAt:     formatted["expires_at"].(string),
			RemainingDays: formatted["remaining_days"].(int),
			IsExpired:     formatted["is_expired"].(bool),
		})
	}

	return &dto.TrashListResponse{Items: result, Total: len(result)}, nil
}

func (s *TrashAppService) Restore(req dto.TrashRestoreRequest) (*dto.TrashRestoreResponse, error) {
	success := make([]string, 0)
	failed := make([]string, 0)

	for _, path := range req.OriginalPaths {
		if err := s.trashDomainSvc.RestoreItem(path); err != nil {
			failed = append(failed, path+": "+err.Error())
		} else {
			success = append(success, path)
		}
	}

	return &dto.TrashRestoreResponse{Success: success, Failed: failed}, nil
}

func (s *TrashAppService) CleanExpired() (*dto.TrashCleanResponse, error) {
	cleaned, err := s.trashDomainSvc.CleanExpiredItems()
	if err != nil {
		return nil, fmt.Errorf("清理过期文件失败: %w", err)
	}

	return &dto.TrashCleanResponse{Cleaned: cleaned}, nil
}
