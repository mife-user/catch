package service

import (
	"catch/internal/domain/entity"
	"catch/internal/domain/repository"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileDomainService struct {
	fileRepo repository.FileRepository
}

func NewFileDomainService(fileRepo repository.FileRepository) *FileDomainService {
	return &FileDomainService{fileRepo: fileRepo}
}

func (s *FileDomainService) SearchFiles(path string, pattern string, fileType entity.FileType, customExts []string, minSize, maxSize int64, modAfter, modBefore string) ([]*entity.FileInfo, []string, error) {
	return s.fileRepo.Search(path, pattern, fileType, customExts, minSize, maxSize, modAfter, modBefore)
}

func (s *FileDomainService) MoveToTrash(filePath string, trashPath string, expireDays int) (*entity.TrashItem, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法获取文件信息: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(trashPath), 0755); err != nil {
		return nil, fmt.Errorf("无法创建回收站目录: %w", err)
	}

	if err := os.Rename(filePath, trashPath); err != nil {
		return nil, fmt.Errorf("无法移动文件到回收站: %w", err)
	}

	return entity.NewTrashItem(filePath, trashPath, info.Size(), expireDays), nil
}

func (s *FileDomainService) RestoreFromTrash(item *entity.TrashItem) error {
	dir := filepath.Dir(item.OriginalPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("无法创建目标目录: %w", err)
	}

	if err := os.Rename(item.TrashPath, item.OriginalPath); err != nil {
		return fmt.Errorf("无法恢复文件: %w", err)
	}

	return nil
}

func (s *FileDomainService) PermanentDelete(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("无法永久删除文件: %w", err)
	}
	return nil
}

func (s *FileDomainService) MoveToSystemTrash(filePath string) error {
	return moveToSystemTrash(filePath)
}

func (s *FileDomainService) GenerateTrashPath(trashDir, originalPath string) string {
	ext := filepath.Ext(originalPath)
	base := strings.TrimSuffix(filepath.Base(originalPath), ext)
	timestamp := time.Now().Format("20060102_150405")
	return filepath.Join(trashDir, fmt.Sprintf("%s_%s%s", base, timestamp, ext))
}

func (s *FileDomainService) GenerateRenamePattern(files []*entity.FileInfo, rule string, params map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	for i, file := range files {
		newName, err := s.applyRenameRule(file, rule, params, i)
		if err != nil {
			return nil, err
		}
		result[file.Path] = filepath.Join(filepath.Dir(file.Path), newName)
	}
	return result, nil
}

func (s *FileDomainService) applyRenameRule(file *entity.FileInfo, rule string, params map[string]string, index int) (string, error) {
	ext := filepath.Ext(file.Name)
	base := strings.TrimSuffix(file.Name, ext)

	switch rule {
	case "prefix":
		prefix := params["prefix"]
		return prefix + base + ext, nil
	case "suffix":
		suffix := params["suffix"]
		return base + suffix + ext, nil
	case "sequence":
		startNum := 1
		if v, ok := params["start"]; ok {
			fmt.Sscanf(v, "%d", &startNum)
		}
		digits := 3
		if v, ok := params["digits"]; ok {
			fmt.Sscanf(v, "%d", &digits)
		}
		seq := index + startNum
		format := fmt.Sprintf("%%0%dd", digits)
		return fmt.Sprintf("%s_"+format+"%s", base, seq, ext), nil
	case "replace":
		old := params["old"]
		newStr := params["new"]
		return strings.ReplaceAll(base, old, newStr) + ext, nil
	case "timestamp":
		ts := time.Now().Format("20060102")
		return base + "_" + ts + ext, nil
	default:
		return "", fmt.Errorf("不支持的重命名规则: %s", rule)
	}
}
