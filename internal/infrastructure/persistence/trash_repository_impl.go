package persistence

import (
	"catch/internal/domain/entity"
	"catch/internal/domain/repository"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type TrashRepositoryImpl struct {
	mu       sync.RWMutex
	filePath string
}

func NewTrashRepository() repository.TrashRepository {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	trashMetaDir := filepath.Join(home, ".catch")

	if err := os.MkdirAll(trashMetaDir, 0755); err != nil {
		exePath, _ := os.Executable()
		fallbackDir := filepath.Join(filepath.Dir(exePath), ".catch")
		if mkErr := os.MkdirAll(fallbackDir, 0755); mkErr != nil {
			fallbackDir = filepath.Join(".", ".catch")
			os.MkdirAll(fallbackDir, 0755)
		}
		trashMetaDir = fallbackDir
	}

	trashMetaPath := filepath.Join(trashMetaDir, "trash.json")
	return &TrashRepositoryImpl{filePath: trashMetaPath}
}

func (r *TrashRepositoryImpl) loadAll() ([]*entity.TrashItem, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*entity.TrashItem{}, nil
		}
		return nil, err
	}

	var items []*entity.TrashItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *TrashRepositoryImpl) saveAll(items []*entity.TrashItem) error {
	configDir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := r.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, r.filePath)
}

func (r *TrashRepositoryImpl) Add(item *entity.TrashItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	items, err := r.loadAll()
	if err != nil {
		return fmt.Errorf("加载回收站数据失败: %w", err)
	}

	items = append(items, item)
	return r.saveAll(items)
}

func (r *TrashRepositoryImpl) List() ([]*entity.TrashItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.loadAll()
}

func (r *TrashRepositoryImpl) FindByOriginalPath(originalPath string) (*entity.TrashItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items, err := r.loadAll()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.OriginalPath == originalPath {
			return item, nil
		}
	}

	return nil, fmt.Errorf("未找到回收站记录: %s", originalPath)
}

func (r *TrashRepositoryImpl) Remove(originalPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	items, err := r.loadAll()
	if err != nil {
		return err
	}

	filtered := make([]*entity.TrashItem, 0, len(items))
	found := false
	for _, item := range items {
		if item.OriginalPath == originalPath {
			found = true
			continue
		}
		filtered = append(filtered, item)
	}

	if !found {
		return fmt.Errorf("未找到回收站记录: %s", originalPath)
	}

	return r.saveAll(filtered)
}

func (r *TrashRepositoryImpl) GetExpiredItems() ([]*entity.TrashItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items, err := r.loadAll()
	if err != nil {
		return nil, err
	}

	expired := make([]*entity.TrashItem, 0)
	for _, item := range items {
		if item.IsExpired() {
			expired = append(expired, item)
		}
	}

	return expired, nil
}

func (r *TrashRepositoryImpl) CleanExpired() (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items, err := r.loadAll()
	if err != nil {
		return 0, err
	}

	filtered := make([]*entity.TrashItem, 0)
	cleaned := 0
	for _, item := range items {
		if item.IsExpired() {
			os.Remove(item.TrashPath)
			cleaned++
		} else {
			filtered = append(filtered, item)
		}
	}

	if cleaned > 0 {
		if err := r.saveAll(filtered); err != nil {
			return 0, err
		}
	}

	return cleaned, nil
}
