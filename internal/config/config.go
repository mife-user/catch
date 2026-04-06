package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 应用程序配置
type Config struct {
	DefaultRecursive    bool     `json:"default_recursive"`
	DefaultPageSize     int      `json:"default_page_size"`
	DefaultContextLines int      `json:"default_context_lines"`
	DefaultMaxGoroutine int      `json:"default_max_goroutine"`
	DefaultMaxFileSize  int64    `json:"default_max_file_size"`
	DefaultMaxMatches   int      `json:"default_max_matches"`
	DefaultExportFormat string   `json:"default_export_format"`
	HistoryMaxEntries   int      `json:"history_max_entries"`
	SkipDirs            []string `json:"skip_dirs"`
	SkipExtensions      []string `json:"skip_extensions"`
	Theme               string   `json:"theme"`
}

// GetDefaultConfig 返回硬编码默认配置
func GetDefaultConfig() *Config {
	return &Config{
		DefaultRecursive:    false,
		DefaultPageSize:     10,
		DefaultContextLines: 0,
		DefaultMaxGoroutine: 10,
		DefaultMaxFileSize:  10 * 1024 * 1024, // 10MB
		DefaultMaxMatches:   100,
		DefaultExportFormat: "json",
		HistoryMaxEntries:   50,
		SkipDirs:            []string{".git", "node_modules", "vendor", ".vscode", ".idea", "__pycache__"},
		SkipExtensions:      []string{".exe", ".dll", ".so", ".dylib", ".zip", ".tar", ".gz"},
		Theme:               "auto",
	}
}

// LoadConfig 加载配置文件
// 查找顺序：当前目录 .catchrc → 用户主目录 .catchrc → 使用硬编码默认值
func LoadConfig() (*Config, error) {
	cfg := GetDefaultConfig()

	// 尝试从当前目录加载
	currentDirConfig, err := loadFromFile(".catchrc")
	if err == nil {
		// 合并配置
		mergeConfig(cfg, currentDirConfig)
		return cfg, nil
	}

	// 尝试从用户主目录加载
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homeConfigPath := filepath.Join(homeDir, ".catchrc")
		homeConfig, err := loadFromFile(homeConfigPath)
		if err == nil {
			mergeConfig(cfg, homeConfig)
			return cfg, nil
		}
	}

	// 返回默认配置
	return cfg, nil
}

// loadFromFile 从指定路径加载配置文件
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &cfg, nil
}

// mergeConfig 合并配置（source 中的值覆盖 target）
func mergeConfig(target, source *Config) {
	if source.DefaultRecursive {
		target.DefaultRecursive = source.DefaultRecursive
	}
	if source.DefaultPageSize > 0 {
		target.DefaultPageSize = source.DefaultPageSize
	}
	if source.DefaultContextLines >= 0 {
		target.DefaultContextLines = source.DefaultContextLines
	}
	if source.DefaultMaxGoroutine > 0 {
		target.DefaultMaxGoroutine = source.DefaultMaxGoroutine
	}
	if source.DefaultMaxFileSize > 0 {
		target.DefaultMaxFileSize = source.DefaultMaxFileSize
	}
	if source.DefaultMaxMatches > 0 {
		target.DefaultMaxMatches = source.DefaultMaxMatches
	}
	if source.DefaultExportFormat != "" {
		target.DefaultExportFormat = source.DefaultExportFormat
	}
	if source.HistoryMaxEntries > 0 {
		target.HistoryMaxEntries = source.HistoryMaxEntries
	}
	if len(source.SkipDirs) > 0 {
		target.SkipDirs = source.SkipDirs
	}
	if len(source.SkipExtensions) > 0 {
		target.SkipExtensions = source.SkipExtensions
	}
	if source.Theme != "" {
		target.Theme = source.Theme
	}
}

// SaveConfig 保存配置到文件
func SaveConfig(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// CreateDefaultConfigFile 生成默认配置文件到指定路径
func CreateDefaultConfigFile(path string) error {
	cfg := GetDefaultConfig()
	return SaveConfig(cfg, path)
}

// GetConfigPath 获取配置文件路径（优先当前目录，其次用户主目录）
func GetConfigPath() string {
	// 检查当前目录
	if _, err := os.Stat(".catchrc"); err == nil {
		absPath, _ := filepath.Abs(".catchrc")
		return absPath
	}

	// 检查用户主目录
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homeConfigPath := filepath.Join(homeDir, ".catchrc")
		if _, err := os.Stat(homeConfigPath); err == nil {
			return homeConfigPath
		}
	}

	return ""
}
