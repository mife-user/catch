package entity

type CleanupRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	FileTypes   []string `json:"file_types"`
	Paths       []string `json:"paths"`
	MinSize     int64    `json:"min_size"`
	MaxSize     int64    `json:"max_size"`
	OlderThan   int      `json:"older_than"`
	Important   bool     `json:"important"`
	BuiltIn     bool     `json:"built_in"`
}

type CleanupResult struct {
	Found    []CleanupFileItem `json:"found"`
	Total    int               `json:"total"`
	TotalSize int64            `json:"total_size"`
}

type CleanupFileItem struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	ModTime   string `json:"mod_time"`
	Extension string `json:"extension"`
	Important bool   `json:"important"`
	RuleName  string `json:"rule_name"`
}

type CleanupExecuteResult struct {
	Cleaned int      `json:"cleaned"`
	Failed  []string `json:"failed"`
	Freed   int64    `json:"freed"`
}

var DefaultCleanupRules = []CleanupRule{
	{
		ID:          "temp_files",
		Name:        "临时文件",
		Description: "清理系统临时文件和缓存",
		Category:    "system",
		FileTypes:   []string{".tmp", ".temp", ".bak", ".cache", ".log"},
		Paths:       []string{},
		OlderThan:   1,
		BuiltIn:     true,
	},
	{
		ID:          "log_files",
		Name:        "日志文件",
		Description: "清理应用程序日志文件",
		Category:    "system",
		FileTypes:   []string{".log"},
		Paths:       []string{},
		OlderThan:   7,
		BuiltIn:     true,
	},
	{
		ID:          "large_files",
		Name:        "大文件",
		Description: "查找超过100MB的文件",
		Category:    "size",
		MinSize:     104857600,
		Paths:       []string{},
		BuiltIn:     true,
	},
	{
		ID:          "qq_cache",
		Name:        "QQ缓存",
		Description: "清理QQ聊天软件的缓存文件",
		Category:    "app_cache",
		FileTypes:   []string{".db", ".sqlite", ".dat"},
		Paths:       []string{},
		BuiltIn:     true,
		Important:   true,
	},
	{
		ID:          "wechat_cache",
		Name:        "微信缓存",
		Description: "清理微信聊天软件的缓存文件",
		Category:    "app_cache",
		FileTypes:   []string{".db", ".sqlite", ".dat"},
		Paths:       []string{},
		BuiltIn:     true,
		Important:   true,
	},
	{
		ID:          "empty_dirs",
		Name:        "空文件夹",
		Description: "查找空文件夹",
		Category:    "structure",
		Paths:       []string{},
		BuiltIn:     true,
	},
}
