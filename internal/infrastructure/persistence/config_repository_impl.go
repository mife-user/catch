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

type ConfigRepositoryImpl struct {
	manager *ConfigManagerAdapter
}

type ConfigManagerAdapter struct {
	mu       sync.RWMutex
	filePath string
}

func NewConfigManagerAdapter() *ConfigManagerAdapter {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	configDir := filepath.Join(home, ".catch")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		exePath, _ := os.Executable()
		fallbackDir := filepath.Join(filepath.Dir(exePath), ".catch")
		if mkErr := os.MkdirAll(fallbackDir, 0755); mkErr != nil {
			fallbackDir = filepath.Join(".", ".catch")
			os.MkdirAll(fallbackDir, 0755)
		}
		configDir = fallbackDir
	}

	configPath := filepath.Join(configDir, "config.json")
	return &ConfigManagerAdapter{filePath: configPath}
}

func NewConfigRepository() repository.ConfigRepository {
	return &ConfigRepositoryImpl{manager: NewConfigManagerAdapter()}
}

func (r *ConfigRepositoryImpl) Load() (*entity.AppConfig, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()

	data, err := os.ReadFile(r.manager.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return entity.DefaultAppConfig(), nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config entity.AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		backupPath := r.manager.filePath + ".bak"
		os.Rename(r.manager.filePath, backupPath)
		return entity.DefaultAppConfig(), fmt.Errorf("配置文件格式错误，已备份至 %s", backupPath)
	}

	return &config, nil
}

func (r *ConfigRepositoryImpl) Save(config *entity.AppConfig) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()

	configDir := filepath.Dir(r.manager.filePath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	tmpFile := r.manager.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	if err := os.Rename(tmpFile, r.manager.filePath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}

func (r *ConfigRepositoryImpl) GetConfigPath() string {
	return r.manager.filePath
}

func (r *ConfigRepositoryImpl) Exists() bool {
	_, err := os.Stat(r.manager.filePath)
	return err == nil
}
