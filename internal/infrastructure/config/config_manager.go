package config

import (
	"catch/internal/domain/entity"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type ConfigManager struct {
	mu       sync.RWMutex
	filePath string
}

func NewConfigManager() *ConfigManager {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	configDir := filepath.Join(home, ".catch")
	configPath := filepath.Join(configDir, "config.json")

	return &ConfigManager{
		filePath: configPath,
	}
}

func (m *ConfigManager) Load() (*entity.AppConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return entity.DefaultAppConfig(), nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config entity.AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		backupPath := m.filePath + ".bak"
		os.Rename(m.filePath, backupPath)
		return entity.DefaultAppConfig(), fmt.Errorf("配置文件格式错误，已备份至 %s", backupPath)
	}

	return &config, nil
}

func (m *ConfigManager) Save(config *entity.AppConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	configDir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	tmpFile := m.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	if err := os.Rename(tmpFile, m.filePath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}

func (m *ConfigManager) GetConfigPath() string {
	return m.filePath
}

func (m *ConfigManager) Exists() bool {
	_, err := os.Stat(m.filePath)
	return err == nil
}

func (m *ConfigManager) EnsureDir() error {
	configDir := filepath.Dir(m.filePath)
	return os.MkdirAll(configDir, 0755)
}
