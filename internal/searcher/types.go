package searcher

// SearchTask 搜索任务
type SearchTask struct {
	FilePath   string
	KeyWord    string
	SearchType bool // true=搜索文件名，false=搜索文件内容
}

// SearchResult 搜索结果
type SearchResult struct {
	FilePath string
	LinNum   int
	Content  string
}

// Config 搜索配置
type Config struct {
	WorkerNum      int // 工作协程数量
	ChanBufferSize int // 通道缓冲区大小
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		WorkerNum:      4,
		ChanBufferSize: 100,
	}
}

// Stats 搜索统计信息
type Stats struct {
	FilesScanned   int64 // 扫描文件数
	LinesScanned   int64 // 扫描行数
	MatchesFound   int64 // 匹配结果数
	ErrorsOccurred int64 // 发生错误数 (包括权限不足跳过的文件)
	FilesSkipped   int64 // 跳过的文件数 (权限不足)
}
