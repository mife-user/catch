package service

import (
	"catch/internal/application/dto"
	"catch/internal/domain/entity"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

type CleanupAppService struct{}

func NewCleanupAppService() *CleanupAppService {
	return &CleanupAppService{}
}

func (s *CleanupAppService) GetRules() *dto.CleanupRulesResponse {
	rules := make([]dto.CleanupRuleItem, 0, len(entity.DefaultCleanupRules))
	for _, r := range entity.DefaultCleanupRules {
		rules = append(rules, dto.CleanupRuleItem{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Category:    r.Category,
			FileTypes:   r.FileTypes,
			OlderThan:   r.OlderThan,
			Important:   r.Important,
			BuiltIn:     r.BuiltIn,
		})
	}
	return &dto.CleanupRulesResponse{Rules: rules}
}

func (s *CleanupAppService) Scan(req dto.CleanupScanRequest, progressCb func(done int, total int)) (*dto.CleanupScanResponse, error) {
	if req.Path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("无法获取用户主目录: %w", err)
		}
		req.Path = home
	}

	selectedRules := make(map[string]*entity.CleanupRule)
	for _, r := range entity.DefaultCleanupRules {
		if slices.Contains(req.Rules, r.ID) {
			selectedRules[r.ID] = &r
		}
	}

	if len(selectedRules) == 0 {
		for i := range entity.DefaultCleanupRules {
			selectedRules[entity.DefaultCleanupRules[i].ID] = &entity.DefaultCleanupRules[i]
		}
	}

	var itemsMu sync.Mutex
	var items []dto.CleanupFileItem
	var totalSize int64

	totalRules := len(selectedRules)
	processedRules := 0

	for _, rule := range selectedRules {
		foundItems := s.scanByRule(req.Path, rule)
		itemsMu.Lock()
		items = append(items, foundItems...)
		for _, item := range foundItems {
			totalSize += item.Size
		}
		itemsMu.Unlock()

		processedRules++
		if progressCb != nil {
			progressCb(processedRules, totalRules)
		}
	}

	return &dto.CleanupScanResponse{
		Items:     items,
		Total:     len(items),
		TotalSize: totalSize,
	}, nil
}

func (s *CleanupAppService) scanByRule(rootPath string, rule *entity.CleanupRule) []dto.CleanupFileItem {
	var items []dto.CleanupFileItem

	now := time.Now()
	cutoff := now.Add(-time.Duration(rule.OlderThan) * 24 * time.Hour)

	if rule.ID == "empty_dirs" {
		s.scanEmptyDirs(rootPath, &items)
		return items
	}

	if rule.ID == "qq_cache" {
		s.scanAppCache(rootPath, "Tencent", "QQ", rule, &items)
		return items
	}

	if rule.ID == "wechat_cache" {
		s.scanAppCache(rootPath, "Tencent", "WeChat", rule, &items)
		return items
	}

	filepath.Walk(rootPath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))

		if len(rule.FileTypes) > 0 {
			if !slices.Contains(rule.FileTypes, ext) {
				return nil
			}
		}

		if rule.MinSize > 0 && info.Size() < rule.MinSize {
			return nil
		}

		if rule.MaxSize > 0 && info.Size() > rule.MaxSize {
			return nil
		}

		if rule.OlderThan > 0 && info.ModTime().After(cutoff) {
			return nil
		}

		items = append(items, dto.CleanupFileItem{
			Name:      info.Name(),
			Path:      currentPath,
			Size:      info.Size(),
			ModTime:   info.ModTime().Format("2006-01-02 15:04:05"),
			Extension: ext,
			Important: rule.Important,
			RuleName:  rule.Name,
		})

		return nil
	})

	return items
}

func (s *CleanupAppService) scanEmptyDirs(rootPath string, items *[]dto.CleanupFileItem) {
	filepath.Walk(rootPath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() {
			return nil
		}

		if currentPath == rootPath {
			return nil
		}

		entries, readErr := os.ReadDir(currentPath)
		if readErr != nil {
			return nil
		}

		if len(entries) == 0 {
			*items = append(*items, dto.CleanupFileItem{
				Name:      info.Name(),
				Path:      currentPath,
				Size:      0,
				ModTime:   info.ModTime().Format("2006-01-02 15:04:05"),
				Extension: "",
				Important: false,
				RuleName:  "空文件夹",
			})
		}

		return nil
	})
}

func (s *CleanupAppService) scanAppCache(_ string, parentDir string, appName string, rule *entity.CleanupRule, items *[]dto.CleanupFileItem) {
	appDataPaths := []string{}
	home, err := os.UserHomeDir()
	if err == nil {
		appDataPaths = append(appDataPaths,
			filepath.Join(home, "AppData", "Roaming", parentDir, appName),
			filepath.Join(home, "AppData", "Local", parentDir, appName),
			filepath.Join(home, "Documents", parentDir, appName),
		)
	}

	for _, appPath := range appDataPaths {
		if _, err := os.Stat(appPath); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(appPath, func(currentPath string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(info.Name()))
			isCache := false
			cacheExts := []string{".db", ".sqlite", ".dat", ".tmp", ".cache", ".jpg", ".png", ".gif", ".mp4", ".amr", ".aud", ".video", ".img"}
			if slices.Contains(cacheExts, ext) {
				isCache = true
			}

			if strings.Contains(strings.ToLower(info.Name()), "cache") ||
				strings.Contains(strings.ToLower(info.Name()), "temp") {
				isCache = true
			}

			if !isCache {
				return nil
			}

			*items = append(*items, dto.CleanupFileItem{
				Name:      info.Name(),
				Path:      currentPath,
				Size:      info.Size(),
				ModTime:   info.ModTime().Format("2006-01-02 15:04:05"),
				Extension: ext,
				Important: rule.Important,
				RuleName:  rule.Name,
			})

			return nil
		})
	}
}

func (s *CleanupAppService) Execute(req dto.CleanupExecuteRequest, progressCb func(done int, total int)) (*dto.CleanupExecuteResponse, error) {
	cleaned := 0
	failed := make([]string, 0)
	var freed int64
	total := len(req.Paths)

	for i, path := range req.Paths {
		info, err := os.Stat(path)
		if err != nil {
			failed = append(failed, path+": "+err.Error())
			if progressCb != nil {
				progressCb(i+1, total)
			}
			continue
		}

		if info.IsDir() {
			if err := os.RemoveAll(path); err != nil {
				failed = append(failed, path+": "+err.Error())
			} else {
				cleaned++
			}
		} else {
			if err := os.Remove(path); err != nil {
				failed = append(failed, path+": "+err.Error())
			} else {
				cleaned++
				freed += info.Size()
			}
		}

		if progressCb != nil {
			progressCb(i+1, total)
		}
	}

	return &dto.CleanupExecuteResponse{
		Cleaned: cleaned,
		Failed:  failed,
		Freed:   freed,
	}, nil
}
