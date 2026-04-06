package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// HistoryEntry 搜索历史记录条目
type HistoryEntry struct {
	ID           int       `json:"id"`
	Keyword      string    `json:"keyword"`
	Path         string    `json:"path"`
	Recursive    bool      `json:"recursive"`
	FileType     string    `json:"file_type"`
	ContextLines int       `json:"context_lines"`
	UseRegex     bool      `json:"use_regex"`
	RegexPattern string    `json:"regex_pattern"`
	SearchMode   string    `json:"search_mode"`
	Keywords     []string  `json:"keywords"`
	Timestamp    time.Time `json:"timestamp"`
}

// HistoryManager 历史记录管理器
type HistoryManager struct {
	entries    []HistoryEntry
	maxEntries int
	filePath   string
	nextID     int
}

// NewHistoryManager 创建历史记录管理器
func NewHistoryManager(maxEntries int) (*HistoryManager, error) {
	hm := &HistoryManager{
		entries:    make([]HistoryEntry, 0),
		maxEntries: maxEntries,
		nextID:     1,
	}

	// 设置历史文件路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %v", err)
	}

	hm.filePath = filepath.Join(homeDir, ".catch_history.json")

	// 加载历史记录
	if err := hm.Load(); err != nil {
		// 如果加载失败，创建新的历史记录
		hm.entries = make([]HistoryEntry, 0)
	}

	return hm, nil
}

// AddEntry 添加搜索记录
func (hm *HistoryManager) AddEntry(entry HistoryEntry) {
	entry.ID = hm.nextID
	entry.Timestamp = time.Now()
	hm.nextID++

	// 添加到开头
	hm.entries = append([]HistoryEntry{entry}, hm.entries...)

	// 限制历史记录数量
	if len(hm.entries) > hm.maxEntries {
		hm.entries = hm.entries[:hm.maxEntries]
	}

	// 自动保存
	hm.Save()
}

// GetRecentEntries 获取最近 N 条记录
func (hm *HistoryManager) GetRecentEntries(n int) []HistoryEntry {
	if n <= 0 || n > len(hm.entries) {
		n = len(hm.entries)
	}
	return hm.entries[:n]
}

// GetEntry 根据 ID 获取记录
func (hm *HistoryManager) GetEntry(id int) (*HistoryEntry, error) {
	for _, entry := range hm.entries {
		if entry.ID == id {
			return &entry, nil
		}
	}
	return nil, fmt.Errorf("未找到 ID 为 %d 的历史记录", id)
}

// Clear 清空历史记录
func (hm *HistoryManager) Clear() {
	hm.entries = make([]HistoryEntry, 0)
	hm.nextID = 1
	hm.Save()
}

// Save 保存历史记录到文件
func (hm *HistoryManager) Save() error {
	data, err := json.MarshalIndent(hm.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化历史记录失败: %v", err)
	}

	if err := os.WriteFile(hm.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入历史记录文件失败: %v", err)
	}

	return nil
}

// Load 从文件加载历史记录
func (hm *HistoryManager) Load() error {
	// 检查文件是否存在
	if _, err := os.Stat(hm.filePath); os.IsNotExist(err) {
		return nil // 文件不存在，不算错误
	}

	data, err := os.ReadFile(hm.filePath)
	if err != nil {
		return fmt.Errorf("读取历史记录文件失败: %v", err)
	}

	if err := json.Unmarshal(data, &hm.entries); err != nil {
		return fmt.Errorf("解析历史记录失败: %v", err)
	}

	// 更新 nextID
	if len(hm.entries) > 0 {
		maxID := 0
		for _, entry := range hm.entries {
			if entry.ID > maxID {
				maxID = entry.ID
			}
		}
		hm.nextID = maxID + 1
	}

	return nil
}

// Count 返回历史记录数量
func (hm *HistoryManager) Count() int {
	return len(hm.entries)
}

// FormatTimestamp 格式化时间戳为可读字符串
func (e *HistoryEntry) FormatTimestamp() string {
	now := time.Now()
	diff := now.Sub(e.Timestamp)

	if diff < time.Minute {
		return "刚刚"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d 分钟前", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d 小时前", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d 天前", days)
	}

	return e.Timestamp.Format("2006-01-02 15:04")
}

// GetDescription 获取搜索描述
func (e *HistoryEntry) GetDescription() string {
	if e.UseRegex {
		return fmt.Sprintf("[正则] %s", e.RegexPattern)
	}

	if e.SearchMode == "multi_and" || e.SearchMode == "multi_or" {
		mode := "AND"
		if e.SearchMode == "multi_or" {
			mode = "OR"
		}
		return fmt.Sprintf("[%s] %s", mode, joinKeywords(e.Keywords))
	}

	return e.Keyword
}

// joinKeywords 连接关键字
func joinKeywords(keywords []string) string {
	if len(keywords) == 0 {
		return ""
	}
	if len(keywords) == 1 {
		return keywords[0]
	}
	if len(keywords) <= 3 {
		result := ""
		for i, kw := range keywords {
			if i > 0 {
				result += ", "
			}
			result += kw
		}
		return result
	}
	return joinKeywords(keywords[:3]) + "..."
}
