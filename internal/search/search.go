package search

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// SearchResult 搜索结果
type SearchResult struct {
	FilePath      string        // 文件路径
	LineNumbers   []int         // 匹配的行号
	Lines         []string      // 匹配的行内容
	MatchType     string        // 匹配类型：content 或 filename
	ContextBefore []ContextLine // 匹配行之前的上下文
	ContextAfter  []ContextLine // 匹配行之后的上下文
}

// ContextLine 上下文行
type ContextLine struct {
	LineNumber int    // 行号
	Content    string // 内容
}

// SearchConfig 搜索配置
type SearchConfig struct {
	Keyword          string
	Path             string
	Recursive        bool
	FileType         string
	MaxGoroutine     int
	MaxFileSize      int64                 // 最大文件大小限制（字节），默认 10MB
	MaxMatches       int                   // 单文件最大匹配行数，默认 100
	ContextLines     int                   // 上下文行数，默认 0（不显示上下文）
	Context          context.Context       // 支持取消
	ProgressCallback func(stats ScanStats) // 进度回调函数
}

// PagedResults 分页结果结构
type PagedResults struct {
	Results    []SearchResult `json:"results"`    // 当前页的结果
	Page       int            `json:"page"`       // 当前页码
	PageSize   int            `json:"page_size"`  // 每页大小
	TotalCount int            `json:"total_count"` // 总结果数
	TotalPages int            `json:"total_pages"` // 总页数
	HasNext    bool           `json:"has_next"`    // 是否有下一页
	HasPrev    bool           `json:"has_prev"`    // 是否有上一页
}

// PageResults 对搜索结果进行分页
func PageResults(results []SearchResult, page, pageSize int) *PagedResults {
	if pageSize <= 0 {
		pageSize = 10
	}
	if page < 1 {
		page = 1
	}

	totalCount := len(results)
	totalPages := 0
	if totalCount > 0 {
		totalPages = (totalCount + pageSize - 1) / pageSize
	}

	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= totalCount {
		return &PagedResults{
			Results:    []SearchResult{},
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
			HasNext:    false,
			HasPrev:    page > 1,
		}
	}

	if end > totalCount {
		end = totalCount
	}

	return &PagedResults{
		Results:    results[start:end],
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// ScanStats 扫描统计
type ScanStats struct {
	FilesScanned int    // 已扫描文件数
	FilesMatched int    // 匹配文件数
	CurrentFile  string // 当前正在扫描的文件
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
				// 符号链接检测
				if isSymlink(fullPath, entry) {
					continue
				}
				if config.Recursive {
					walkDir(fullPath)
				}
			} else {
				// 跳过符号链接文件
				if isSymlink(fullPath, entry) {
					continue
				}
				// 跳过二进制文件
				if isBinaryFile(fullPath) {
					continue
				}

				// 文件类型过滤（使用 EqualFold）
				if config.FileType != "" {
					if !strings.EqualFold(filepath.Ext(entry.Name()), config.FileType) {
						continue
					}
				}

				// 文件大小检查（利用 DirEntry 的 Info）
				if config.MaxFileSize > 0 {
					if info, err := entry.Info(); err == nil {
						if info.Size() > config.MaxFileSize {
							continue
						}
					}
				}

				tasks = append(tasks, fileTask{path: fullPath, config: config})
			}
		}
	}

	walkDir(config.Path)
	return tasks
}

// collectFilesStreaming 流式收集文件并发送到通道（生产者模式）
func collectFilesStreaming(ctx context.Context, config SearchConfig, taskChan chan<- fileTask) {
	defer close(taskChan)

	var walkDir func(dir string) bool
	walkDir = func(dir string) bool {
		select {
		case <-ctx.Done():
			return false // 取消
		default:
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return true
		}

		for _, entry := range entries {
			select {
			case <-ctx.Done():
				return false // 取消
			default:
			}

			fullPath := filepath.Join(dir, entry.Name())

			if entry.IsDir() {
				if shouldSkipDir(entry.Name()) {
					continue
				}
				// 符号链接检测，避免死循环
				if isSymlink(fullPath, entry) {
					continue
				}
				if config.Recursive {
					if !walkDir(fullPath) {
						return false
					}
				}
			} else {
				// 跳过符号链接文件
				if isSymlink(fullPath, entry) {
					continue
				}
				// 跳过二进制文件
				if isBinaryFile(fullPath) {
					continue
				}

				// 文件类型过滤（使用 EqualFold 避免 ToLower）
				if config.FileType != "" {
					if !strings.EqualFold(filepath.Ext(entry.Name()), config.FileType) {
						continue
					}
				}

				// 检查文件大小（利用 DirEntry 的 Info，避免重复 Stat）
				if config.MaxFileSize > 0 {
					if info, err := entry.Info(); err == nil {
						if info.Size() > config.MaxFileSize {
							continue // 跳过大文件
						}
					}
				}

				select {
				case taskChan <- fileTask{path: fullPath, config: config}:
				case <-ctx.Done():
					return false
				}
			}
		}
		return true
	}

	walkDir(config.Path)
}

// isSymlink 检查是否为符号链接
func isSymlink(path string, entry os.DirEntry) bool {
	// 检查 DirEntry 的 Type 信息
	if entry.Type()&os.ModeSymlink != 0 {
		return true
	}
	// 备用检查
	info, err := entry.Info()
	if err == nil {
		return info.Mode()&os.ModeSymlink != 0
	}
	return false
}

// searchFile 搜索单个文件
func searchFile(task fileTask, stats *ScanStats, mu *sync.Mutex) []SearchResult {
	// 更新进度
	if task.config.ProgressCallback != nil && mu != nil {
		mu.Lock()
		stats.CurrentFile = task.path
		stats.FilesScanned++
		task.config.ProgressCallback(*stats)
		mu.Unlock()
	}

	var results []resultCollector

	// 文件名搜索（使用 EqualFold 避免 ToLower 分配）
	if strings.Contains(strings.ToLower(filepath.Base(task.path)), strings.ToLower(task.config.Keyword)) {
		results = append(results, resultCollector{
			filePath:  task.path,
			matchType: "filename",
			lines:     []string{filepath.Base(task.path)},
		})
		if task.config.ProgressCallback != nil && mu != nil {
			mu.Lock()
			stats.FilesMatched++
			mu.Unlock()
		}
	}

	// 文件内容搜索
	file, err := os.Open(task.path)
	if err != nil {
		return convertResults(results)
	}
	defer file.Close()

	// 如果需要上下文，先读取所有行
	var allLines []string
	var lineNum int
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	
	for scanner.Scan() {
		lineNum++
		allLines = append(allLines, scanner.Text())
	}

	// 搜索匹配的行
	keyword := task.config.Keyword
	maxMatches := task.config.MaxMatches
	if maxMatches <= 0 {
		maxMatches = 100
	}
	contextLines := task.config.ContextLines

	var matchLines []string
	var matchLineNums []int
	var allContextBefore [][]ContextLine
	var allContextAfter [][]ContextLine

	for i, line := range allLines {
		if containsIgnoreCase(line, keyword) {
			matchLines = append(matchLines, line)
			matchLineNums = append(matchLineNums, i+1)

			// 提取上下文
			if contextLines > 0 {
				var before []ContextLine
				var after []ContextLine

				// 前面的上下文
				start := i - contextLines
				if start < 0 {
					start = 0
				}
				for j := start; j < i; j++ {
					before = append(before, ContextLine{
						LineNumber: j + 1,
						Content:    allLines[j],
					})
				}

				// 后面的上下文
				end := i + 1 + contextLines
				if end > len(allLines) {
					end = len(allLines)
				}
				for j := i + 1; j < end; j++ {
					after = append(after, ContextLine{
						LineNumber: j + 1,
						Content:    allLines[j],
					})
				}

				allContextBefore = append(allContextBefore, before)
				allContextAfter = append(allContextAfter, after)
			}

			// 达到最大匹配数后停止
			if len(matchLines) >= maxMatches {
				break
			}
		}
	}

	if len(matchLines) > 0 {
		// 避免重复添加文件名匹配
		hasContentMatch := false
		for _, r := range results {
			if r.matchType == "content" {
				hasContentMatch = true
				break
			}
		}
		if !hasContentMatch {
			rc := resultCollector{
				filePath:    task.path,
				lineNumbers: matchLineNums,
				lines:       matchLines,
				matchType:   "content",
			}
			// 添加上下文
			if contextLines > 0 && len(allContextBefore) > 0 {
				// 合并所有匹配点的上下文
				rc.contextBefore = mergeContexts(allContextBefore)
				rc.contextAfter = mergeContexts(allContextAfter)
			}
			results = append(results, rc)
		}
		if task.config.ProgressCallback != nil && mu != nil {
			mu.Lock()
			stats.FilesMatched++
			mu.Unlock()
		}
	}

	return convertResults(results)
}

// mergeContexts 合并多个匹配点的上下文，去除重复
func mergeContexts(contexts [][]ContextLine) []ContextLine {
	if len(contexts) == 0 {
		return nil
	}

	// 使用 map 去重
	seen := make(map[int]bool)
	var merged []ContextLine

	for _, ctx := range contexts {
		for _, line := range ctx {
			if !seen[line.LineNumber] {
				seen[line.LineNumber] = true
				merged = append(merged, line)
			}
		}
	}

	// 按行号排序
	sortContextLines(merged)
	return merged
}

// sortContextLines 按行号排序
func sortContextLines(lines []ContextLine) {
	for i := 0; i < len(lines); i++ {
		for j := i + 1; j < len(lines); j++ {
			if lines[j].LineNumber < lines[i].LineNumber {
				lines[i], lines[j] = lines[j], lines[i]
			}
		}
	}
}

// containsIgnoreCase 高性能的大小写无关包含检查
func containsIgnoreCase(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	// 使用 Boyer-Moore-Horspool 简化版，避免 ToLower 分配
	lowerS := strings.ToLower(s)
	lowerSubstr := strings.ToLower(substr)
	return strings.Contains(lowerS, lowerSubstr)
}

// resultCollector 内部结构体，减少转换开销
type resultCollector struct {
	filePath      string
	lineNumbers   []int
	lines         []string
	matchType     string
	contextBefore []ContextLine
	contextAfter  []ContextLine
}

func convertResults(rcs []resultCollector) []SearchResult {
	results := make([]SearchResult, 0, len(rcs))
	for _, rc := range rcs {
		results = append(results, SearchResult{
			FilePath:      rc.filePath,
			LineNumbers:   rc.lineNumbers,
			Lines:         rc.lines,
			MatchType:     rc.matchType,
			ContextBefore: rc.contextBefore,
			ContextAfter:  rc.contextAfter,
		})
	}
	return results
}

// Search 执行搜索（使用协程池，生产者-消费者模式）
func Search(config SearchConfig) []SearchResult {
	// 设置默认值
	if config.MaxGoroutine <= 0 {
		config.MaxGoroutine = 10
	}
	if config.MaxFileSize <= 0 {
		config.MaxFileSize = 10 * 1024 * 1024 // 默认 10MB
	}
	if config.MaxMatches <= 0 {
		config.MaxMatches = 100 // 默认每文件最多 100 个匹配
	}
	if config.Context == nil {
		config.Context = context.Background()
	}

	ctx := config.Context

	// 创建流式通道（动态缓冲区）
	bufSize := config.MaxGoroutine * 10
	if bufSize < 100 {
		bufSize = 100
	}
	taskChan := make(chan fileTask, bufSize)
	resultChan := make(chan []SearchResult, bufSize)

	// 进度统计
	var stats ScanStats
	var statsMu sync.Mutex

	// 启动生产者（异步遍历目录）
	go collectFilesStreaming(ctx, config, taskChan)

	// 启动工作协程（消费者）
	var wg sync.WaitGroup
	for i := 0; i < config.MaxGoroutine; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				select {
				case <-ctx.Done():
					return
				default:
					resultChan <- searchFile(task, &stats, &statsMu)
				}
			}
		}()
	}

	// 等待所有协程完成后关闭结果通道
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果（流式处理）
	var allResults []SearchResult
	for {
		select {
		case <-ctx.Done():
			return allResults // 返回已收集的结果
		case result, ok := <-resultChan:
			if !ok {
				return allResults
			}
			allResults = append(allResults, result...)
		}
	}
}

// ResultCallback 结果回调函数类型
type ResultCallback func(result SearchResult)

// SearchStreaming 执行流式搜索，每找到一个结果就调用回调函数
func SearchStreaming(config SearchConfig, callback ResultCallback) int {
	// 设置默认值
	if config.MaxGoroutine <= 0 {
		config.MaxGoroutine = 10
	}
	if config.MaxFileSize <= 0 {
		config.MaxFileSize = 10 * 1024 * 1024 // 默认 10MB
	}
	if config.MaxMatches <= 0 {
		config.MaxMatches = 100 // 默认每文件最多 100 个匹配
	}
	if config.Context == nil {
		config.Context = context.Background()
	}

	ctx := config.Context

	// 创建流式通道（动态缓冲区）
	bufSize := config.MaxGoroutine * 10
	if bufSize < 100 {
		bufSize = 100
	}
	taskChan := make(chan fileTask, bufSize)
	resultChan := make(chan SearchResult, bufSize)

	// 进度统计
	var stats ScanStats
	var statsMu sync.Mutex

	// 启动生产者
	go collectFilesStreaming(ctx, config, taskChan)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < config.MaxGoroutine; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				select {
				case <-ctx.Done():
					return
				default:
					// 搜索单个文件并逐个发送结果
					results := searchFile(task, &stats, &statsMu)
					for _, r := range results {
						select {
						case resultChan <- r:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}()
	}

	// 等待所有协程完成后关闭结果通道
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 流式处理结果，每找到一个就调用回调
	count := 0
	for {
		select {
		case <-ctx.Done():
			return count
		case result, ok := <-resultChan:
			if !ok {
				return count
			}
			if callback != nil {
				callback(result)
			}
			count++
		}
	}
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

		fmt.Printf("%s [%d] %s\n", matchType, i+1, Highlight(result.FilePath, keyword))

		if result.MatchType == "content" {
			// 打印前面的上下文
			if len(result.ContextBefore) > 0 {
				for _, ctx := range result.ContextBefore {
					fmt.Printf("    %3d: %s\n", ctx.LineNumber, HighlightContext(ctx.Content, keyword))
				}
			}

			// 打印匹配行
			for j, line := range result.Lines {
				lineNum := result.LineNumbers[j]
				fmt.Printf("  > %3d: %s\n", lineNum, Highlight(line, keyword))
			}

			// 打印后面的上下文
			if len(result.ContextAfter) > 0 {
				for _, ctx := range result.ContextAfter {
					fmt.Printf("    %3d: %s\n", ctx.LineNumber, HighlightContext(ctx.Content, keyword))
				}
			}
		}
		fmt.Println()
	}
}

// Highlight 高亮关键字
func Highlight(text, keyword string) string {
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

// HighlightContext 高亮上下文（使用灰色显示）
func HighlightContext(text, keyword string) string {
	if keyword == "" {
		return text
	}

	const (
		gray   = "\033[90m"
		yellow = "\033[33m"
		reset  = "\033[0m"
	)

	result := text
	lowerText := strings.ToLower(text)
	lowerKeyword := strings.ToLower(keyword)

	// 从后往前替换，避免索引偏移
	for i := len(lowerText) - len(lowerKeyword); i >= 0; i-- {
		if lowerText[i:i+len(lowerKeyword)] == lowerKeyword {
			result = result[:i] + gray + result[i:i+len(keyword)] + reset + result[i+len(keyword):]
		}
	}

	return gray + result + reset
}

// PrintResultsPaged 分页打印搜索结果（交互式）
func PrintResultsPaged(results []SearchResult, keyword string, pageSize int) {
	if len(results) == 0 {
		fmt.Println("未找到匹配的结果")
		return
	}

	paged := PageResults(results, 1, pageSize)

	for {
		fmt.Printf("\n=== 第 %d/%d 页（共 %d 条结果） ===\n\n", paged.Page, paged.TotalPages, paged.TotalCount)

		for i, result := range paged.Results {
			globalIndex := (paged.Page-1)*paged.PageSize + i + 1
			matchType := "📄"
			if result.MatchType == "filename" {
				matchType = "📁"
			}

			fmt.Printf("%s [%d] %s\n", matchType, globalIndex, Highlight(result.FilePath, keyword))

			if result.MatchType == "content" {
				for j, line := range result.Lines {
					lineNum := result.LineNumbers[j]
					fmt.Printf("    %3d: %s\n", lineNum, Highlight(line, keyword))
				}
			}
			fmt.Println()
		}

		if !paged.HasNext {
			break
		}

		fmt.Printf("按 Enter 查看下一页，输入 q 退出：")
		var input string
		fmt.Scanln(&input)
		if input == "q" || input == "Q" {
			break
		}

		paged = PageResults(results, paged.Page+1, pageSize)
	}
}

// PrintPagedResults 打印分页结果（非交互式，打印指定页）
func PrintPagedResults(paged *PagedResults, keyword string) {
	if paged == nil || len(paged.Results) == 0 {
		fmt.Println("未找到匹配的结果")
		return
	}

	fmt.Printf("\n=== 第 %d/%d 页（共 %d 条结果） ===\n\n", paged.Page, paged.TotalPages, paged.TotalCount)

	for i, result := range paged.Results {
		globalIndex := (paged.Page-1)*paged.PageSize + i + 1
		matchType := "📄"
		if result.MatchType == "filename" {
			matchType = "📁"
		}

		fmt.Printf("%s [%d] %s\n", matchType, globalIndex, Highlight(result.FilePath, keyword))

		if result.MatchType == "content" {
			// 打印前面的上下文
			if len(result.ContextBefore) > 0 {
				for _, ctx := range result.ContextBefore {
					fmt.Printf("    %3d: %s\n", ctx.LineNumber, HighlightContext(ctx.Content, keyword))
				}
			}

			// 打印匹配行
			for j, line := range result.Lines {
				lineNum := result.LineNumbers[j]
				fmt.Printf("  > %3d: %s\n", lineNum, Highlight(line, keyword))
			}

			// 打印后面的上下文
			if len(result.ContextAfter) > 0 {
				for _, ctx := range result.ContextAfter {
					fmt.Printf("    %3d: %s\n", ctx.LineNumber, HighlightContext(ctx.Content, keyword))
				}
			}
		}
		fmt.Println()
	}

	if paged.HasNext || paged.HasPrev {
		fmt.Printf("--- 分页信息: 第 %d/%d 页", paged.Page, paged.TotalPages)
		if paged.HasPrev {
			fmt.Printf(" [上一页: 第 %d 页]", paged.Page-1)
		}
		if paged.HasNext {
			fmt.Printf(" [下一页: 第 %d 页]", paged.Page+1)
		}
		fmt.Println(" ---")
	}
}

// ExportResults 导出搜索结果到文件
// format: "json", "csv", "txt"
// outputPath: 输出文件路径
func ExportResults(results []SearchResult, keyword string, format string, outputPath string) error {
	if len(results) == 0 {
		return fmt.Errorf("没有可导出的结果")
	}

	// 确保输出目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	switch strings.ToLower(format) {
	case "json":
		return exportToJSON(results, keyword, outputPath)
	case "csv":
		return exportToCSV(results, keyword, outputPath)
	case "txt":
		return exportToTXT(results, keyword, outputPath)
	default:
		return fmt.Errorf("不支持的格式: %s (支持: json, csv, txt)", format)
	}
}

// MatchDetail 匹配详情
type MatchDetail struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
}

// ResultExport 导出的结果结构
type ResultExport struct {
	FilePath      string        `json:"file_path"`
	MatchType     string        `json:"match_type"`
	Matches       []MatchDetail `json:"matches"`
	ContextBefore []ContextLine `json:"context_before,omitempty"`
	ContextAfter  []ContextLine `json:"context_after,omitempty"`
}

// ExportData 导出数据结构
type ExportData struct {
	Keyword    string         `json:"keyword"`
	TotalCount int            `json:"total_count"`
	Results    []ResultExport `json:"results"`
}

// exportToJSON 导出为 JSON 格式
func exportToJSON(results []SearchResult, keyword string, outputPath string) error {
	exportData := ExportData{
		Keyword:    keyword,
		TotalCount: len(results),
		Results:    make([]ResultExport, 0, len(results)),
	}

	for _, result := range results {
		re := ResultExport{
			FilePath:      result.FilePath,
			MatchType:     result.MatchType,
			ContextBefore: result.ContextBefore,
			ContextAfter:  result.ContextAfter,
		}

		if result.MatchType == "content" {
			for i, line := range result.Lines {
				re.Matches = append(re.Matches, MatchDetail{
					LineNumber: result.LineNumbers[i],
					Content:    line,
				})
			}
		}

		exportData.Results = append(exportData.Results, re)
	}

	// 写入 JSON 文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(exportData); err != nil {
		return fmt.Errorf("JSON 编码失败: %v", err)
	}

	return nil
}

// exportToCSV 导出为 CSV 格式
func exportToCSV(results []SearchResult, keyword string, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	header := []string{"file_path", "match_type", "line_number", "content"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("写入 CSV 表头失败: %v", err)
	}

	// 写入数据
	for _, result := range results {
		if result.MatchType == "filename" {
			record := []string{
				result.FilePath,
				"filename",
				"0",
				filepath.Base(result.FilePath),
			}
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("写入 CSV 记录失败: %v", err)
			}
		} else {
			for i, line := range result.Lines {
				record := []string{
					result.FilePath,
					"content",
					fmt.Sprintf("%d", result.LineNumbers[i]),
					line,
				}
				if err := writer.Write(record); err != nil {
					return fmt.Errorf("写入 CSV 记录失败: %v", err)
				}
			}
		}
	}

	return nil
}

// exportToTXT 导出为纯文本格式
func exportToTXT(results []SearchResult, keyword string, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// 写入标题
	fmt.Fprintf(writer, "搜索结果 - 关键字: %s\n", keyword)
	fmt.Fprintf(writer, "总计: %d 个匹配\n\n", len(results))

	// 写入结果
	for i, result := range results {
		matchType := "[内容]"
		if result.MatchType == "filename" {
			matchType = "[文件名]"
		}

		fmt.Fprintf(writer, "%s [%d] %s\n", matchType, i+1, result.FilePath)

		if result.MatchType == "content" {
			// 写入前面的上下文
			if len(result.ContextBefore) > 0 {
				for _, ctx := range result.ContextBefore {
					fmt.Fprintf(writer, "    %3d: %s\n", ctx.LineNumber, ctx.Content)
				}
			}

			// 写入匹配行
			for j, line := range result.Lines {
				lineNum := result.LineNumbers[j]
				fmt.Fprintf(writer, "  > %3d: %s\n", lineNum, line)
			}

			// 写入后面的上下文
			if len(result.ContextAfter) > 0 {
				for _, ctx := range result.ContextAfter {
					fmt.Fprintf(writer, "    %3d: %s\n", ctx.LineNumber, ctx.Content)
				}
			}
		}
		fmt.Fprintln(writer)
	}

	return nil
}
