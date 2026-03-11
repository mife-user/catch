package search

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// SearchResult 搜索结果
type SearchResult struct {
	FilePath    string   // 文件路径
	LineNumbers []int    // 匹配的行号
	Lines       []string // 匹配的行内容
	MatchType   string   // 匹配类型：content 或 filename
}

// SearchConfig 搜索配置
type SearchConfig struct {
	Keyword      string
	Path         string
	Recursive    bool
	FileType     string
	MaxGoroutine int
}

// fileTask 文件搜索任务
type fileTask struct {
	path   string
	config SearchConfig
}

// skipExtensions 跳过的文件扩展名
var skipExtensions = map[string]bool{
	".exe": true, ".dll": true, ".so": true, ".dylib": true,
	".bin": true, ".obj": true, ".o": true, ".a": true,
	".lib": true, ".pyc": true, ".pyo": true, ".class": true,
	".zip": true, ".tar": true, ".gz": true, ".rar": true,
	".7z": true, ".pdf": true, ".doc": true, ".docx": true,
	".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".bmp": true, ".ico": true, ".svg": true, ".webp": true,
	".mp3": true, ".mp4": true, ".avi": true, ".mkv": true,
	".wav": true, ".flac": true, ".wma": true, ".mid": true,
	".db": true, ".sqlite": true, ".sqlite3": true, ".mdb": true,
}

// shouldSkipDir 判断是否跳过目录
func shouldSkipDir(name string) bool {
	skipDirs := []string{".git", "node_modules", "vendor", ".vscode", ".idea", "__pycache__", ".qwen"}
	for _, skip := range skipDirs {
		if name == skip {
			return true
		}
	}
	return strings.HasPrefix(name, ".")
}

// shouldSkipFile 判断是否跳过文件
func shouldSkipFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return skipExtensions[ext]
}

// isBinaryFile 检查文件是否为二进制文件
func isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if skipExtensions[ext] {
		return true
	}

	// 检查文件名是否包含常见二进制文件标识
	name := strings.ToLower(filepath.Base(path))
	binaryIndicators := []string{".min.", ".bundle", ".map"}
	for _, indicator := range binaryIndicators {
		if strings.Contains(name, indicator) {
			return true
		}
	}

	return false
}

// collectFiles 收集所有需要搜索的文件
func collectFiles(config SearchConfig) []fileTask {
	var tasks []fileTask

	var walkDir func(dir string)
	walkDir = func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}

		for _, entry := range entries {
			fullPath := filepath.Join(dir, entry.Name())

			if entry.IsDir() {
				if shouldSkipDir(entry.Name()) {
					continue
				}
				if config.Recursive {
					walkDir(fullPath)
				}
			} else {
				// 跳过二进制文件
				if isBinaryFile(fullPath) {
					continue
				}

				// 文件类型过滤
				if config.FileType != "" {
					if !strings.HasSuffix(strings.ToLower(entry.Name()), strings.ToLower(config.FileType)) {
						continue
					}
				}
				tasks = append(tasks, fileTask{path: fullPath, config: config})
			}
		}
	}

	walkDir(config.Path)
	return tasks
}

// searchFile 搜索单个文件
func searchFile(task fileTask) []SearchResult {
	var results []SearchResult

	// 文件名搜索
	if strings.Contains(strings.ToLower(filepath.Base(task.path)), strings.ToLower(task.config.Keyword)) {
		results = append(results, SearchResult{
			FilePath:  task.path,
			MatchType: "filename",
			Lines:     []string{filepath.Base(task.path)},
		})
	}

	// 文件内容搜索
	file, err := os.Open(task.path)
	if err != nil {
		return results
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var matchLines []string
	var matchLineNums []int

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), strings.ToLower(task.config.Keyword)) {
			matchLines = append(matchLines, line)
			matchLineNums = append(matchLineNums, lineNum)
		}
	}

	if len(matchLines) > 0 {
		results = append(results, SearchResult{
			FilePath:    task.path,
			LineNumbers: matchLineNums,
			Lines:       matchLines,
			MatchType:   "content",
		})
	}

	return results
}

// Search 执行搜索（使用协程池）
func Search(config SearchConfig) []SearchResult {
	if config.MaxGoroutine <= 0 {
		config.MaxGoroutine = 10
	}

	// 收集所有文件任务
	tasks := collectFiles(config)
	if len(tasks) == 0 {
		return []SearchResult{}
	}

	// 创建任务通道
	taskChan := make(chan fileTask, len(tasks))
	resultChan := make(chan []SearchResult, len(tasks))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < config.MaxGoroutine; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				resultChan <- searchFile(task)
			}
		}()
	}

	// 发送所有任务
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	// 等待所有协程完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var allResults []SearchResult
	for result := range resultChan {
		allResults = append(allResults, result...)
	}

	return allResults
}

// PrintResults 打印搜索结果（带高亮）
func PrintResults(results []SearchResult, keyword string) {
	if len(results) == 0 {
		fmt.Println("未找到匹配的结果")
		return
	}

	fmt.Printf("找到 %d 个匹配结果:\n\n", len(results))

	for i, result := range results {
		matchType := "📄"
		if result.MatchType == "filename" {
			matchType = "📁"
		}

		fmt.Printf("%s [%d] %s\n", matchType, i+1, highlight(result.FilePath, keyword))

		if result.MatchType == "content" {
			for j, line := range result.Lines {
				lineNum := result.LineNumbers[j]
				fmt.Printf("    %3d: %s\n", lineNum, highlight(line, keyword))
			}
		}
		fmt.Println()
	}
}

// highlight 高亮关键字
func highlight(text, keyword string) string {
	if keyword == "" {
		return text
	}

	// Windows 终端颜色代码
	const (
		yellow = "\033[33m"
		reset  = "\033[0m"
	)

	result := text
	lowerText := strings.ToLower(text)
	lowerKeyword := strings.ToLower(keyword)

	// 从后往前替换，避免索引偏移
	for i := len(lowerText) - len(lowerKeyword); i >= 0; i-- {
		if lowerText[i:i+len(lowerKeyword)] == lowerKeyword {
			result = result[:i] + yellow + result[i:i+len(keyword)] + reset + result[i+len(keyword):]
		}
	}

	return result
}

// PrintResultsPaged 分页打印搜索结果
func PrintResultsPaged(results []SearchResult, keyword string, pageSize int) {
	if len(results) == 0 {
		fmt.Println("未找到匹配的结果")
		return
	}

	totalPages := (len(results) + pageSize - 1) / pageSize
	currentPage := 1

	for {
		start := (currentPage - 1) * pageSize
		end := start + pageSize
		if end > len(results) {
			end = len(results)
		}

		fmt.Printf("\n=== 第 %d/%d 页 ===\n\n", currentPage, totalPages)

		for i := start; i < end; i++ {
			result := results[i]
			matchType := "📄"
			if result.MatchType == "filename" {
				matchType = "📁"
			}

			fmt.Printf("%s [%d] %s\n", matchType, i+1, highlight(result.FilePath, keyword))

			if result.MatchType == "content" {
				for j, line := range result.Lines {
					lineNum := result.LineNumbers[j]
					fmt.Printf("    %3d: %s\n", lineNum, highlight(line, keyword))
				}
			}
			fmt.Println()
		}

		if currentPage >= totalPages {
			break
		}

		fmt.Printf("按 Enter 继续，输入 q 退出：")
		var input string
		fmt.Scanln(&input)
		if input == "q" || input == "Q" {
			break
		}
		currentPage++
	}
}
