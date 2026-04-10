package service

import (
	"catch/internal/domain/entity"
	"catch/internal/domain/repository"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type TrashDomainService struct {
	trashRepo  repository.TrashRepository
	configRepo repository.ConfigRepository
}

func NewTrashDomainService(trashRepo repository.TrashRepository, configRepo repository.ConfigRepository) *TrashDomainService {
	return &TrashDomainService{trashRepo: trashRepo, configRepo: configRepo}
}

func (s *TrashDomainService) CleanExpiredItems() (int, error) {
	expired, err := s.trashRepo.GetExpiredItems()
	if err != nil {
		return 0, fmt.Errorf("获取过期文件列表失败: %w", err)
	}

	cleaned := 0
	for _, item := range expired {
		if err := os.Remove(item.TrashPath); err != nil {
			continue
		}
		if err := s.trashRepo.Remove(item.OriginalPath); err != nil {
			continue
		}
		cleaned++
	}

	return cleaned, nil
}

func (s *TrashDomainService) GetTrashDir() (string, error) {
	config, err := s.configRepo.Load()
	if err != nil {
		return "", err
	}

	trashPath := config.Trash.Path
	if trashPath == "~/.catch-trash" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("无法获取用户主目录: %w", err)
		}
		trashPath = filepath.Join(home, ".catch-trash")
	}

	if err := os.MkdirAll(trashPath, 0755); err != nil {
		return "", fmt.Errorf("无法创建回收站目录: %w", err)
	}

	return trashPath, nil
}

func (s *TrashDomainService) StartupCleanup() error {
	cleaned, err := s.CleanExpiredItems()
	if err != nil {
		return err
	}

	if cleaned > 0 {
		fmt.Printf("已清理 %d 个过期文件\n", cleaned)
	}

	return nil
}

func (s *TrashDomainService) RestoreItem(originalPath string) error {
	item, err := s.trashRepo.FindByOriginalPath(originalPath)
	if err != nil {
		return fmt.Errorf("未找到回收站记录: %w", err)
	}

	if _, err := os.Stat(item.TrashPath); os.IsNotExist(err) {
		return fmt.Errorf("回收站文件不存在: %s", item.TrashPath)
	}

	dir := filepath.Dir(item.OriginalPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("无法创建目标目录: %w", err)
	}

	if err := os.Rename(item.TrashPath, item.OriginalPath); err != nil {
		return fmt.Errorf("无法恢复文件: %w", err)
	}

	if err := s.trashRepo.Remove(originalPath); err != nil {
		return fmt.Errorf("无法删除回收站记录: %w", err)
	}

	return nil
}

func (s *TrashDomainService) GetExpireDays() int {
	config, err := s.configRepo.Load()
	if err != nil {
		return 7
	}
	return config.Trash.ExpireDays
}

func (s *TrashDomainService) FormatTrashItem(item *entity.TrashItem) map[string]interface{} {
	remaining := time.Until(item.ExpiresAt)
	if remaining < 0 {
		remaining = 0
	}
	return map[string]interface{}{
		"original_path":    item.OriginalPath,
		"trash_path":       item.TrashPath,
		"file_name":        item.FileName,
		"file_size":        item.FileSize,
		"deleted_at":       item.DeletedAt.Format("2006-01-02 15:04:05"),
		"expires_at":       item.ExpiresAt.Format("2006-01-02 15:04:05"),
		"remaining_days":   int(remaining.Hours() / 24),
		"is_expired":       item.IsExpired(),
	}
}
