package cli

import (
	"bufio"
	"catch/internal/config"
	"catch/internal/history"
	"catch/internal/search"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

// Writer 输出接口
type Writer interface {
	WriteString(s string) (int, error)
	Flush() error
}

// SimpleWriter 简单写入器
type SimpleWriter struct {
	Writer io.Writer
}

func (w *SimpleWriter) WriteString(s string) (int, error) {
	return fmt.Fprint(w.Writer, s)
}

func (w *SimpleWriter) Flush() error {
	if f, ok := w.Writer.(*os.File); ok {
		return f.Sync()
	}
	return nil
}

// MenuItem 菜单项
type MenuItem struct {
	Title    string
	Action   func()
	Shortcut string
}

// RunInteractive 运行交互式界面
func RunInteractive() {
	reader := bufio.NewReader(os.Stdin)
	writer := &SimpleWriter{Writer: os.Stdout}

	// 初始化终端（如果需要）
	if term.IsTerminal(int(os.Stdin.Fd())) {
		// 终端模式
		runTerminalUI(reader, writer)
	} else {
		// 非终端模式
		runSimpleUI(reader, writer)
	}
}

func runTerminalUI(reader *bufio.Reader, writer *SimpleWriter) {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		writer.WriteString(fmt.Sprintf("⚠️  加载配置失败，使用默认配置: %v\n", err))
		writer.Flush()
		cfg = config.GetDefaultConfig()
	}

	// 加载历史记录
	historyMgr, err := history.NewHistoryManager(cfg.HistoryMaxEntries)
	if err != nil {
		writer.WriteString(fmt.Sprintf("⚠️  加载历史记录失败: %v\n", err))
		writer.Flush()
		historyMgr = nil
	}

	menuItems := []MenuItem{
		{Title: "🔍 搜索文件内容", Action: func() { searchContent(reader, writer, cfg, historyMgr) }, Shortcut: "1"},
		{Title: "📁 搜索文件名", Action: func() { searchFilename(reader, writer, cfg, historyMgr) }, Shortcut: "2"},
		{Title: "⚙️  高级搜索", Action: func() { advancedSearch(reader, writer, cfg, historyMgr) }, Shortcut: "3"},
		{Title: "🔎 正则表达式搜索", Action: func() { regexSearch(reader, writer, cfg, historyMgr) }, Shortcut: "4"},
		{Title: "🔗 多关键字搜索", Action: func() { multiKeywordSearch(reader, writer, cfg, historyMgr) }, Shortcut: "5"},
		{Title: "📜 搜索历史", Action: func() { searchHistory(reader, writer, historyMgr, cfg) }, Shortcut: "6"},
		{Title: "⚙️  配置管理", Action: func() { configManager(reader, writer, cfg) }, Shortcut: "7"},
		{Title: "➕ 添加到环境变量", Action: func() { AddToPath(writer) }, Shortcut: "8"},
		{Title: "❌ 退出", Action: func() {}, Shortcut: "q"},
	}

	// 显示欢迎信息
	writer.WriteString("╔════════════════════════════════════════════╗\n")
	writer.WriteString("║     🎯 Catch - 文件搜索工具                ║\n")
	writer.WriteString("║     高性能本地文件内容搜索                 ║\n")
	writer.WriteString("╚════════════════════════════════════════════╝\n")
	writer.WriteString("\n")
	writer.Flush()

	for {
		printMenu(writer, menuItems)
		writer.WriteString("\n请选择功能 (1-8 或 q): ")
		writer.Flush()

		input := readLine(reader)
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "q" || input == "quit" || input == "exit" {
			writer.WriteString("👋 再见!\n")
			writer.Flush()
			break
		}

		// 处理数字选择
		selected, err := strconv.Atoi(input)
		if err != nil || selected < 1 || selected > len(menuItems) {
			writer.WriteString("❌ 无效选择，请重试\n\n")
			writer.Flush()
			continue
		}

		// 操作完成后显示分隔线
		writer.WriteString("\n" + strings.Repeat("─", 50) + "\n\n")
		writer.Flush()
		menuItems[selected-1].Action()
	}
}

func runSimpleUI(reader *bufio.Reader, writer *SimpleWriter) {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		writer.WriteString(fmt.Sprintf("⚠️  加载配置失败，使用默认配置: %v\n", err))
		writer.Flush()
		cfg = config.GetDefaultConfig()
	}

	// 加载历史记录
	historyMgr, err := history.NewHistoryManager(cfg.HistoryMaxEntries)
	if err != nil {
		writer.WriteString(fmt.Sprintf("⚠️  加载历史记录失败: %v\n", err))
		writer.Flush()
		historyMgr = nil
	}

	writer.WriteString("=== Catch - 文件搜索工具 ===\n\n")
	writer.WriteString("1. 搜索文件内容\n")
	writer.WriteString("2. 搜索文件名\n")
	writer.WriteString("3. 高级搜索\n")
	writer.WriteString("4. 添加到环境变量\n")
	writer.WriteString("q. 退出\n\n")
	writer.Flush()

	for {
		writer.WriteString("请选择：")
		writer.Flush()

		input := readLine(reader)
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "q" || input == "quit" || input == "exit" {
			writer.WriteString("👋 再见!\n")
			writer.Flush()
			break
		}

		switch input {
		case "1":
			searchContent(reader, writer, cfg, historyMgr)
		case "2":
			searchFilename(reader, writer, cfg, historyMgr)
		case "3":
			advancedSearch(reader, writer, cfg, historyMgr)
		case "4":
			AddToPath(writer)
		default:
			writer.WriteString("❌ 无效选择\n")
			writer.Flush()
		}
	}
}

// readLine 读取一行输入，处理各种换行符
func readLine(reader *bufio.Reader) string {
	line, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	// 移除各种换行符
	line = strings.TrimRight(line, "\r\n")
	return line
}

func printMenu(writer *SimpleWriter, items []MenuItem) {
	// 移除清屏，改为直接打印分隔线
	writer.WriteString("┌────────────────────────────────────────────┐\n")
	writer.WriteString("│  🎯 Catch - 文件搜索工具                   │\n")
	writer.WriteString("├────────────────────────────────────────────┤\n")

	for _, item := range items {
		// 计算显示宽度：中文和 Emoji 算 2 个字符，英文算 1 个
		title := item.Title
		displayWidth := getDisplayWidth(title)
		shortcutWidth := len(item.Shortcut) + 2 // [1]
		// 总宽度 44（中文字符算 2）
		padding := 44 - displayWidth - shortcutWidth
		if padding < 0 {
			padding = 0
		}
		writer.WriteString(fmt.Sprintf("│  [%s] %s%s│\n", item.Shortcut, title, strings.Repeat(" ", padding)))
	}

	writer.WriteString("└────────────────────────────────────────────┘\n")
	writer.Flush()
}

// getDisplayWidth 计算字符串的显示宽度（中文和 Emoji 算 2 个字符）
func getDisplayWidth(s string) int {
	width := 0
	for _, r := range s {
		switch {
		case r >= 0x4E00 && r <= 0x9FFF:
			// 中文字符 (CJK Unified Ideographs)
			width += 2
		case r >= 0x3400 && r <= 0x4DBF:
			// CJK Unified Ideographs Extension A
			width += 2
		case r >= 0xF900 && r <= 0xFAFF:
			// CJK Compatibility Ideographs
			width += 2
		case r >= 0x3000 && r <= 0x303F:
			// CJK 符号和标点
			width += 2
		case r >= 0xFF00 && r <= 0xFFEF:
			// 全角字符
			width += 2
		case r >= 0x1F600 && r <= 0x1F64F:
			// Emoticons (表情符号)
			width += 2
		case r >= 0x1F300 && r <= 0x1F5FF:
			// Miscellaneous Symbols and Pictographs
			width += 2
		case r >= 0x1F680 && r <= 0x1F6FF:
			// Transport and Map Symbols
			width += 2
		case r >= 0x1F1E0 && r <= 0x1F1FF:
			// Regional Indicator Symbols (国旗)
			width += 2
		case r >= 0x2600 && r <= 0x26FF:
			// Miscellaneous Symbols
			width += 2
		case r >= 0x2700 && r <= 0x27BF:
			// Dingbats
			width += 2
		case r >= 0xFE00 && r <= 0xFE0F:
			// Variation Selectors (忽略)
			width += 0
		case r >= 0x200D:
			// Zero Width Joiner (忽略)
			width += 0
		default:
			// ASCII 字符和其他 Unicode 字符
			width += 1
		}
	}
	return width
}

func searchContent(reader *bufio.Reader, writer *SimpleWriter, cfg *config.Config, historyMgr *history.HistoryManager) {
	writer.WriteString("请输入搜索关键字：")
	writer.Flush()

	keyword := readLine(reader)
	keyword = strings.TrimSpace(keyword)

	if keyword == "" {
		writer.WriteString("❌ 关键字不能为空\n")
		writer.Flush()
		pause(writer)
		return
	}

	writer.WriteString(fmt.Sprintf("是否递归搜索子目录？(y/n, 默认 %v): ", cfg.DefaultRecursive))
	writer.Flush()

	recInput := readLine(reader)
	recursive := cfg.DefaultRecursive
	if strings.TrimSpace(strings.ToLower(recInput)) == "y" {
		recursive = true
	} else if strings.TrimSpace(strings.ToLower(recInput)) == "n" {
		recursive = false
	}

	// 验证路径
	searchPath := "."
	if _, err := os.Stat(searchPath); err != nil {
		writer.WriteString(fmt.Sprintf("❌ 路径不存在：%s\n", searchPath))
		writer.Flush()
		pause(writer)
		return
	}

	writer.WriteString("是否使用分页显示？(y/n, 默认 y): ")
	writer.Flush()

	pagedInput := readLine(reader)
	usePaged := true
	if strings.TrimSpace(strings.ToLower(pagedInput)) == "n" {
		usePaged = false
	}

	pageSize := cfg.DefaultPageSize
	if usePaged {
		writer.WriteString(fmt.Sprintf("每页显示条数 (默认 %d): ", cfg.DefaultPageSize))
		writer.Flush()

		pageSizeInput := readLine(reader)
		if size, err := strconv.Atoi(strings.TrimSpace(pageSizeInput)); err == nil && size > 0 {
			pageSize = size
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}

	writer.WriteString(fmt.Sprintf("显示上下文行数 (默认 %d, 不显示为 0): ", cfg.DefaultContextLines))
	writer.Flush()

	contextLinesInput := readLine(reader)
	contextLines := cfg.DefaultContextLines
	if strings.TrimSpace(contextLinesInput) != "" {
		if val, err := strconv.Atoi(strings.TrimSpace(contextLinesInput)); err == nil {
			contextLines = val
		}
	}
	if contextLines < 0 {
		contextLines = 0
	}

	maxGoroutines := cfg.DefaultMaxGoroutine
	if maxGoroutines <= 0 {
		maxGoroutines = 10
	}

	writer.WriteString("是否导出结果？(y/n): ")
	writer.Flush()

	exportInput := readLine(reader)
	exportResults := strings.TrimSpace(strings.ToLower(exportInput)) == "y"

	exportFormat := ""
	exportPath := ""
	if exportResults {
		writer.WriteString("导出格式 (json/csv/txt): ")
		writer.Flush()
		exportFormat = strings.TrimSpace(strings.ToLower(readLine(reader)))
		if exportFormat == "" {
			exportFormat = "json"
		}

		writer.WriteString(fmt.Sprintf("导出路径 (默认 results.%s): ", exportFormat))
		writer.Flush()
		exportPath = strings.TrimSpace(readLine(reader))
		if exportPath == "" {
			exportPath = fmt.Sprintf("results.%s", exportFormat)
		}
	}

	// 创建进度条
	progressBar := NewProgressBar(writer, 0, 30) // 未知总数，使用旋转动画

	// 创建带超时的 context（60秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	startTime := time.Now()
	lastStats := search.ScanStats{}

	config := search.SearchConfig{
		Keyword:      keyword,
		Path:         searchPath,
		Recursive:    recursive,
		MaxGoroutine: maxGoroutines,
		ContextLines: contextLines,
		Context:      ctx,
		ProgressCallback: func(stats search.ScanStats) {
			lastStats = stats
			// 每 10 个文件更新一次进度条（更频繁）
			if stats.FilesScanned%10 == 0 {
				progressBar.Update(stats.FilesScanned, stats)
			}
		},
	}

	// 收集所有结果
	results := search.Search(config)

	// 完成进度条
	progressBar.Finish(lastStats)

	elapsed := time.Since(startTime)
	totalCount := len(results)

	if totalCount == 0 {
		writer.WriteString("未找到匹配的结果\n")
		writer.Flush()
		pause(writer)
		return
	}

	writer.WriteString(fmt.Sprintf("✅ 找到 %d 个匹配结果，耗时 %v\n", totalCount, elapsed.Round(time.Millisecond)))
	writer.Flush()

	// 导出结果
	if exportResults && exportPath != "" {
		writer.WriteString(fmt.Sprintf("📤 正在导出结果到 %s...\n", exportPath))
		writer.Flush()

		if err := search.ExportResults(results, keyword, exportFormat, exportPath); err != nil {
			writer.WriteString(fmt.Sprintf("❌ 导出失败: %v\n", err))
			writer.Flush()
		} else {
			writer.WriteString(fmt.Sprintf("✅ 结果已导出到: %s\n", exportPath))
			writer.Flush()
		}
	}

	if usePaged {
		// 分页显示
		displayPagedResults(writer, results, keyword, pageSize)
	} else {
		// 直接显示所有结果
		search.PrintResults(results, keyword)
	}

	writer.Flush()
	pause(writer)
}

func searchFilename(reader *bufio.Reader, writer *SimpleWriter, cfg *config.Config, historyMgr *history.HistoryManager) {
	writer.WriteString("请输入文件名关键字：")
	writer.Flush()

	keyword := readLine(reader)
	keyword = strings.TrimSpace(keyword)

	if keyword == "" {
		writer.WriteString("❌ 关键字不能为空\n")
		writer.Flush()
		pause(writer)
		return
	}

	writer.WriteString(fmt.Sprintf("是否递归搜索子目录？(y/n, 默认 %v): ", cfg.DefaultRecursive))
	writer.Flush()

	recInput := readLine(reader)
	recursive := cfg.DefaultRecursive
	if strings.TrimSpace(strings.ToLower(recInput)) == "y" {
		recursive = true
	} else if strings.TrimSpace(strings.ToLower(recInput)) == "n" {
		recursive = false
	}

	// 验证路径
	searchPath := "."
	if _, err := os.Stat(searchPath); err != nil {
		writer.WriteString(fmt.Sprintf("❌ 路径不存在：%s\n", searchPath))
		writer.Flush()
		pause(writer)
		return
	}

	maxGoroutines := cfg.DefaultMaxGoroutine
	if maxGoroutines <= 0 {
		maxGoroutines = 10
	}

	// 创建带超时的 context（60秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	startTime := time.Now()

	config := search.SearchConfig{
		Keyword:      keyword,
		Path:         searchPath,
		Recursive:    recursive,
		MaxGoroutine: maxGoroutines,
		Context:      ctx,
		ProgressCallback: func(stats search.ScanStats) {
			if stats.FilesScanned%50 == 0 {
				elapsed := time.Since(startTime)
				speed := float64(stats.FilesScanned) / elapsed.Seconds()
				writer.WriteString(fmt.Sprintf("\r📊 进度: 已扫描 %d 个文件，匹配 %d 个，速度 %.0f 文件/秒",
					stats.FilesScanned, stats.FilesMatched, speed))
				writer.Flush()
			}
		},
	}

	writer.WriteString("\n🔍 正在搜索...\n")
	writer.Flush()

	// 使用流式搜索，只输出文件名匹配的结果
	index := 0
	_ = search.SearchStreaming(config, func(result search.SearchResult) {
		if result.MatchType == "filename" {
			index++
			// 清除进度行
			writer.WriteString("\r" + strings.Repeat(" ", 80) + "\r")
			printStreamingResult(writer, result, keyword, index)
		}
	})

	// 清除最后的进度行
	writer.WriteString("\r" + strings.Repeat(" ", 80) + "\r")

	elapsed := time.Since(startTime)
	if index == 0 {
		writer.WriteString("未找到匹配的文件\n")
	} else {
		writer.WriteString(fmt.Sprintf("✅ 找到 %d 个匹配文件，耗时 %v\n", index, elapsed.Round(time.Millisecond)))
	}
	writer.Flush()
	pause(writer)
}

func advancedSearch(reader *bufio.Reader, writer *SimpleWriter, cfg *config.Config, historyMgr *history.HistoryManager) {
	writer.WriteString("请输入搜索关键字：")
	writer.Flush()

	keyword := readLine(reader)
	keyword = strings.TrimSpace(keyword)

	if keyword == "" {
		writer.WriteString("❌ 关键字不能为空\n")
		writer.Flush()
		pause(writer)
		return
	}

	writer.WriteString("搜索路径 (默认为当前目录): ")
	writer.Flush()

	path := readLine(reader)
	path = strings.TrimSpace(path)
	if path == "" {
		path = "."
	}

	// 验证路径
	if _, err := os.Stat(path); err != nil {
		writer.WriteString(fmt.Sprintf("❌ 路径不存在：%s\n", path))
		writer.Flush()
		pause(writer)
		return
	}

	writer.WriteString(fmt.Sprintf("是否递归搜索子目录？(y/n, 默认 %v): ", cfg.DefaultRecursive))
	writer.Flush()

	recInput := readLine(reader)
	recursive := cfg.DefaultRecursive
	if strings.TrimSpace(strings.ToLower(recInput)) == "y" {
		recursive = true
	} else if strings.TrimSpace(strings.ToLower(recInput)) == "n" {
		recursive = false
	}

	writer.WriteString("文件类型过滤 (如：.go,.txt，留空表示全部): ")
	writer.Flush()

	fileType := readLine(reader)
	fileType = strings.TrimSpace(fileType)

	maxGoroutines := cfg.DefaultMaxGoroutine
	if maxGoroutines <= 0 {
		maxGoroutines = 10
	}

	writer.WriteString("是否使用分页显示？(y/n, 默认 y): ")
	writer.Flush()

	pagedInput := readLine(reader)
	usePaged := true
	if strings.TrimSpace(strings.ToLower(pagedInput)) == "n" {
		usePaged = false
	}

	pageSize := cfg.DefaultPageSize
	if usePaged {
		writer.WriteString(fmt.Sprintf("每页显示条数 (默认 %d): ", cfg.DefaultPageSize))
		writer.Flush()

		pageSizeInput := readLine(reader)
		if size, err := strconv.Atoi(strings.TrimSpace(pageSizeInput)); err == nil && size > 0 {
			pageSize = size
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}

	writer.WriteString(fmt.Sprintf("显示上下文行数 (默认 %d, 不显示为 0): ", cfg.DefaultContextLines))
	writer.Flush()

	contextLinesInput := readLine(reader)
	contextLines := cfg.DefaultContextLines
	if strings.TrimSpace(contextLinesInput) != "" {
		if val, err := strconv.Atoi(strings.TrimSpace(contextLinesInput)); err == nil {
			contextLines = val
		}
	}
	if contextLines < 0 {
		contextLines = 0
	}

	writer.WriteString("是否导出结果？(y/n): ")
	writer.Flush()

	exportInput := readLine(reader)
	exportResults := strings.TrimSpace(strings.ToLower(exportInput)) == "y"

	exportFormat := ""
	exportPath := ""
	if exportResults {
		writer.WriteString("导出格式 (json/csv/txt): ")
		writer.Flush()
		exportFormat = strings.TrimSpace(strings.ToLower(readLine(reader)))
		if exportFormat == "" {
			exportFormat = "json"
		}

		writer.WriteString(fmt.Sprintf("导出路径 (默认 results.%s): ", exportFormat))
		writer.Flush()
		exportPath = strings.TrimSpace(readLine(reader))
		if exportPath == "" {
			exportPath = fmt.Sprintf("results.%s", exportFormat)
		}
	}

	writer.WriteString("是否加载 .gitignore 和 .catchignore 文件？(y/n): ")
	writer.Flush()

	ignoreInput := readLine(reader)
	loadGitignore := strings.TrimSpace(strings.ToLower(ignoreInput)) == "y"

	// 创建带超时的 context（120秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	startTime := time.Now()

	config := search.SearchConfig{
		Keyword:      keyword,
		Path:         path,
		Recursive:    recursive,
		FileType:     fileType,
		MaxGoroutine: maxGoroutines,
		ContextLines: contextLines,
		Context:      ctx,
		LoadGitignore: loadGitignore,
		ProgressCallback: func(stats search.ScanStats) {
			if stats.FilesScanned%50 == 0 {
				elapsed := time.Since(startTime)
				speed := float64(stats.FilesScanned) / elapsed.Seconds()
				writer.WriteString(fmt.Sprintf("\r📊 进度: 已扫描 %d 个文件，匹配 %d 个，速度 %.0f 文件/秒",
					stats.FilesScanned, stats.FilesMatched, speed))
				writer.Flush()
			}
		},
	}

	writer.WriteString("\n🔍 正在搜索...\n")
	writer.Flush()

	// 收集所有结果
	results := search.Search(config)

	// 清除最后的进度行
	writer.WriteString("\r" + strings.Repeat(" ", 80) + "\r")

	elapsed := time.Since(startTime)
	totalCount := len(results)

	if totalCount == 0 {
		writer.WriteString("未找到匹配的结果\n")
		writer.Flush()
		pause(writer)
		return
	}

	writer.WriteString(fmt.Sprintf("✅ 找到 %d 个匹配结果，耗时 %v\n", totalCount, elapsed.Round(time.Millisecond)))
	writer.Flush()

	// 导出结果
	if exportResults && exportPath != "" {
		writer.WriteString(fmt.Sprintf("📤 正在导出结果到 %s...\n", exportPath))
		writer.Flush()

		if err := search.ExportResults(results, keyword, exportFormat, exportPath); err != nil {
			writer.WriteString(fmt.Sprintf("❌ 导出失败: %v\n", err))
			writer.Flush()
		} else {
			writer.WriteString(fmt.Sprintf("✅ 结果已导出到: %s\n", exportPath))
			writer.Flush()
		}
	}

	if usePaged {
		// 分页显示
		displayPagedResults(writer, results, keyword, pageSize)
	} else {
		// 直接显示所有结果
		search.PrintResults(results, keyword)
	}

	writer.Flush()
	pause(writer)
}

func pause(writer *SimpleWriter) {
	writer.WriteString("\n按 Enter 键继续...")
	writer.Flush()
	readLine(bufio.NewReader(os.Stdin))
}

// printStreamingResult 流式打印单个搜索结果
func printStreamingResult(writer *SimpleWriter, result search.SearchResult, keyword string, index int) {
	matchType := "📄"
	if result.MatchType == "filename" {
		matchType = "📁"
	}

	writer.WriteString(fmt.Sprintf("%s [%d] %s\n", matchType, index, search.Highlight(result.FilePath, keyword)))

	if result.MatchType == "content" {
		for j, line := range result.Lines {
			lineNum := result.LineNumbers[j]
			writer.WriteString(fmt.Sprintf("    %3d: %s\n", lineNum, search.Highlight(line, keyword)))
		}
	}
	writer.WriteString("\n")
	writer.Flush()
}

// displayPagedResults 交互式分页显示结果
func displayPagedResults(writer *SimpleWriter, results []search.SearchResult, keyword string, pageSize int) {
	if len(results) == 0 {
		writer.WriteString("未找到匹配的结果\n")
		writer.Flush()
		return
	}

	paged := search.PageResults(results, 1, pageSize)
	reader := bufio.NewReader(os.Stdin)

	for {
		// 打印当前页
		search.PrintPagedResults(paged, keyword)
		writer.Flush()

		// 如果没有更多页，退出
		if !paged.HasNext && !paged.HasPrev {
			break
		}

		// 显示导航
		writer.WriteString("\n导航: [n] 下一页, [p] 上一页, [页码] 跳转, [q] 退出: ")
		writer.Flush()

		input := readLine(reader)
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "q" || input == "quit" || input == "exit" {
			break
		}

		newPage := paged.Page

		switch input {
		case "n", "next":
			if paged.HasNext {
				newPage = paged.Page + 1
			} else {
				writer.WriteString("⚠️  已经是最后一页\n")
				writer.Flush()
				continue
			}
		case "p", "prev", "previous":
			if paged.HasPrev {
				newPage = paged.Page - 1
			} else {
				writer.WriteString("⚠️  已经是第一页\n")
				writer.Flush()
				continue
			}
		default:
			// 尝试解析页码
			if pageNum, err := strconv.Atoi(input); err == nil && pageNum > 0 {
				if pageNum <= paged.TotalPages {
					newPage = pageNum
				} else {
					writer.WriteString(fmt.Sprintf("⚠️  页码超出范围 (1-%d)\n", paged.TotalPages))
					writer.Flush()
					continue
				}
			} else {
				writer.WriteString("❌ 无效输入\n")
				writer.Flush()
				continue
			}
		}

		// 获取新页
		paged = search.PageResults(results, newPage, pageSize)
		
		// 清屏效果（可选）
		writer.WriteString("\n")
		writer.Flush()
	}
}

// AddToPath 添加到系统环境变量
func AddToPath(writer *SimpleWriter) {
	// 获取可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		writer.WriteString("❌ 无法获取可执行文件路径\n")
		writer.Flush()
		return
	}

	execDir := filepath.Dir(execPath)

	writer.WriteString("📍 检测到以下路径需要添加到环境变量:\n")
	writer.WriteString(fmt.Sprintf("   %s\n\n", execDir))
	writer.Flush()

	switch runtime.GOOS {
	case "windows":
		addToPathWindows(execDir, writer)
	case "darwin", "linux":
		addToPathUnix(execDir, writer)
	default:
		writer.WriteString("⚠️  不支持的操作系统\n")
		writer.Flush()
		return
	}
}

func addToPathWindows(path string, writer *SimpleWriter) {
	writer.WriteString("🪟 Windows 系统\n\n")
	writer.WriteString("⚠️  请手动将以下路径添加到系统环境变量 PATH:\n")
	writer.WriteString(fmt.Sprintf("   %s\n\n", path))
	writer.Flush()

	writer.WriteString("方法 1: 使用图形界面 (推荐)\n")
	writer.WriteString("   1. 右键点击'此电脑' -> '属性'\n")
	writer.WriteString("   2. 点击'高级系统设置'\n")
	writer.WriteString("   3. 点击'环境变量'\n")
	writer.WriteString("   4. 在'系统变量'中找到'Path'\n")
	writer.WriteString("   5. 点击'编辑' -> '新建'\n")
	writer.WriteString(fmt.Sprintf("   6. 添加路径：%s\n", path))
	writer.WriteString("   7. 点击'确定'保存\n\n")
	writer.Flush()

	writer.WriteString("方法 2: 使用 PowerShell (需要管理员权限)\n")
	writer.WriteString("   以管理员身份运行 PowerShell，执行:\n")
	writer.WriteString("   $currentPath = [Environment]::GetEnvironmentVariable('Path', 'Machine')\n")
	writer.WriteString(fmt.Sprintf("   [Environment]::SetEnvironmentVariable('Path', \"$currentPath;%s\", 'Machine')\n", path))
	writer.WriteString("\n")
	writer.WriteString("💡 添加完成后，请重启终端或重新登录使环境变量生效\n")
	writer.Flush()
}

func addToPathUnix(path string, writer *SimpleWriter) {
	writer.WriteString("🐧 macOS/Linux 系统\n\n")
	writer.Flush()

	shellConfig := "~/.bashrc"
	if os.Getenv("SHELL") != "" && strings.Contains(os.Getenv("SHELL"), "zsh") {
		shellConfig = "~/.zshrc"
	}

	writer.WriteString(fmt.Sprintf("请将以下行添加到您的 shell 配置文件 (%s):\n", shellConfig))
	writer.WriteString(fmt.Sprintf("   export PATH=\"$PATH:%s\"\n", path))
	writer.WriteString("\n")
	writer.WriteString("然后执行:\n")
	writer.WriteString("   source ~/.bashrc  # 或 source ~/.zshrc\n\n")
	writer.Flush()

	// 尝试自动添加
	writer.WriteString("尝试自动添加到 ~/.bashrc...\n")
	writer.Flush()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		writer.WriteString("⚠️  无法获取用户主目录\n")
		writer.Flush()
		return
	}

	bashrcPath := filepath.Join(homeDir, ".bashrc")
	exportLine := fmt.Sprintf("export PATH=\"$PATH:%s\"\n", path)

	// 检查是否已存在
	content, _ := os.ReadFile(bashrcPath)
	if strings.Contains(string(content), path) {
		writer.WriteString("✅ 路径已存在于 ~/.bashrc\n")
		writer.Flush()
		return
	}

	// 追加到文件
	f, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		writer.WriteString("⚠️  无法写入 ~/.bashrc，请手动添加\n")
		writer.Flush()
		return
	}
	defer f.Close()

	_, err = f.WriteString(exportLine)
	if err != nil {
		writer.WriteString("⚠️  写入失败，请手动添加\n")
		writer.Flush()
		return
	}

	writer.WriteString("✅ 已添加到 ~/.bashrc，请执行 'source ~/.bashrc' 生效\n")
	writer.Flush()
}

// regexSearch 正则表达式搜索
func regexSearch(reader *bufio.Reader, writer *SimpleWriter, cfg *config.Config, historyMgr *history.HistoryManager) {
	writer.WriteString("请输入正则表达式：")
	writer.Flush()

	pattern := readLine(reader)
	pattern = strings.TrimSpace(pattern)

	if pattern == "" {
		writer.WriteString("❌ 正则表达式不能为空\n")
		writer.Flush()
		pause(writer)
		return
	}

	// 编译正则表达式
	regexPattern, err := search.CompileRegex(pattern)
	if err != nil {
		writer.WriteString(fmt.Sprintf("❌ 正则表达式编译失败: %v\n", err))
		writer.Flush()
		pause(writer)
		return
	}

	writer.WriteString(fmt.Sprintf("是否递归搜索子目录？(y/n, 默认 %v): ", cfg.DefaultRecursive))
	writer.Flush()

	recInput := readLine(reader)
	recursive := cfg.DefaultRecursive
	if strings.TrimSpace(strings.ToLower(recInput)) == "y" {
		recursive = true
	} else if strings.TrimSpace(strings.ToLower(recInput)) == "n" {
		recursive = false
	}

	// 验证路径
	searchPath := "."
	if _, err := os.Stat(searchPath); err != nil {
		writer.WriteString(fmt.Sprintf("❌ 路径不存在：%s\n", searchPath))
		writer.Flush()
		pause(writer)
		return
	}

	writer.WriteString("是否使用分页显示？(y/n, 默认 y): ")
	writer.Flush()

	pagedInput := readLine(reader)
	usePaged := true
	if strings.TrimSpace(strings.ToLower(pagedInput)) == "n" {
		usePaged = false
	}

	pageSize := cfg.DefaultPageSize
	if usePaged {
		writer.WriteString(fmt.Sprintf("每页显示条数 (默认 %d): ", cfg.DefaultPageSize))
		writer.Flush()

		pageSizeInput := readLine(reader)
		if size, err := strconv.Atoi(strings.TrimSpace(pageSizeInput)); err == nil && size > 0 {
			pageSize = size
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}

	writer.WriteString(fmt.Sprintf("显示上下文行数 (默认 %d, 不显示为 0): ", cfg.DefaultContextLines))
	writer.Flush()

	contextLinesInput := readLine(reader)
	contextLines := cfg.DefaultContextLines
	if strings.TrimSpace(contextLinesInput) != "" {
		if val, err := strconv.Atoi(strings.TrimSpace(contextLinesInput)); err == nil {
			contextLines = val
		}
	}
	if contextLines < 0 {
		contextLines = 0
	}

	maxGoroutines := cfg.DefaultMaxGoroutine
	if maxGoroutines <= 0 {
		maxGoroutines = 10
	}

	// 创建带超时的 context（60秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	startTime := time.Now()

	config := search.SearchConfig{
		Keyword:      pattern,
		Path:         searchPath,
		Recursive:    recursive,
		MaxGoroutine: maxGoroutines,
		ContextLines: contextLines,
		Context:      ctx,
		UseRegex:     true,
		RegexPattern: regexPattern,
		ProgressCallback: func(stats search.ScanStats) {
			if stats.FilesScanned%50 == 0 {
				elapsed := time.Since(startTime)
				speed := float64(stats.FilesScanned) / elapsed.Seconds()
				writer.WriteString(fmt.Sprintf("\r📊 进度: 已扫描 %d 个文件，匹配 %d 个，速度 %.0f 文件/秒",
					stats.FilesScanned, stats.FilesMatched, speed))
				writer.Flush()
			}
		},
	}

	writer.WriteString("\n🔍 正在搜索...\n")
	writer.Flush()

	// 收集所有结果
	results := search.Search(config)

	// 清除最后的进度行
	writer.WriteString("\r" + strings.Repeat(" ", 80) + "\r")

	elapsed := time.Since(startTime)
	totalCount := len(results)

	if totalCount == 0 {
		writer.WriteString("未找到匹配的结果\n")
		writer.Flush()
		pause(writer)
		return
	}

	writer.WriteString(fmt.Sprintf("✅ 找到 %d 个匹配结果，耗时 %v\n", totalCount, elapsed.Round(time.Millisecond)))
	writer.Flush()

	if usePaged {
		displayPagedResults(writer, results, pattern, pageSize)
	} else {
		search.PrintResults(results, pattern)
	}

	writer.Flush()
	pause(writer)
}

// multiKeywordSearch 多关键字搜索
func multiKeywordSearch(reader *bufio.Reader, writer *SimpleWriter, cfg *config.Config, historyMgr *history.HistoryManager) {
	writer.WriteString("请输入多个关键字（用逗号分隔）：")
	writer.Flush()

	input := readLine(reader)
	input = strings.TrimSpace(input)

	if input == "" {
		writer.WriteString("❌ 关键字不能为空\n")
		writer.Flush()
		pause(writer)
		return
	}

	// 解析关键字
	keywords := strings.Split(input, ",")
	for i, kw := range keywords {
		keywords[i] = strings.TrimSpace(kw)
	}
	// 过滤空关键字
	var validKeywords []string
	for _, kw := range keywords {
		if kw != "" {
			validKeywords = append(validKeywords, kw)
		}
	}

	if len(validKeywords) == 0 {
		writer.WriteString("❌ 至少需要一个有效关键字\n")
		writer.Flush()
		pause(writer)
		return
	}

	// 选择搜索模式
	writer.WriteString("请选择搜索模式 (1: AND - 所有关键字都必须匹配, 2: OR - 任一关键字匹配即可): ")
	writer.Flush()

	modeInput := strings.TrimSpace(readLine(reader))
	searchMode := "multi_and"
	modeDesc := "AND"
	if modeInput == "2" {
		searchMode = "multi_or"
		modeDesc = "OR"
	}

	writer.WriteString(fmt.Sprintf("是否递归搜索子目录？(y/n, 默认 %v): ", cfg.DefaultRecursive))
	writer.Flush()

	recInput := readLine(reader)
	recursive := cfg.DefaultRecursive
	if strings.TrimSpace(strings.ToLower(recInput)) == "y" {
		recursive = true
	} else if strings.TrimSpace(strings.ToLower(recInput)) == "n" {
		recursive = false
	}

	// 验证路径
	searchPath := "."
	if _, err := os.Stat(searchPath); err != nil {
		writer.WriteString(fmt.Sprintf("❌ 路径不存在：%s\n", searchPath))
		writer.Flush()
		pause(writer)
		return
	}

	writer.WriteString("是否使用分页显示？(y/n, 默认 y): ")
	writer.Flush()

	pagedInput := readLine(reader)
	usePaged := true
	if strings.TrimSpace(strings.ToLower(pagedInput)) == "n" {
		usePaged = false
	}

	pageSize := cfg.DefaultPageSize
	if usePaged {
		writer.WriteString(fmt.Sprintf("每页显示条数 (默认 %d): ", cfg.DefaultPageSize))
		writer.Flush()

		pageSizeInput := readLine(reader)
		if size, err := strconv.Atoi(strings.TrimSpace(pageSizeInput)); err == nil && size > 0 {
			pageSize = size
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}

	writer.WriteString(fmt.Sprintf("显示上下文行数 (默认 %d, 不显示为 0): ", cfg.DefaultContextLines))
	writer.Flush()

	contextLinesInput := readLine(reader)
	contextLines := cfg.DefaultContextLines
	if strings.TrimSpace(contextLinesInput) != "" {
		if val, err := strconv.Atoi(strings.TrimSpace(contextLinesInput)); err == nil {
			contextLines = val
		}
	}
	if contextLines < 0 {
		contextLines = 0
	}

	maxGoroutines := cfg.DefaultMaxGoroutine
	if maxGoroutines <= 0 {
		maxGoroutines = 10
	}

	// 创建带超时的 context（60秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	startTime := time.Now()

	// 组合关键字用于显示
	combinedKeyword := strings.Join(validKeywords, fmt.Sprintf(" [%s] ", modeDesc))

	config := search.SearchConfig{
		Keyword:      combinedKeyword,
		Path:         searchPath,
		Recursive:    recursive,
		MaxGoroutine: maxGoroutines,
		ContextLines: contextLines,
		Context:      ctx,
		SearchMode:   searchMode,
		Keywords:     validKeywords,
		ProgressCallback: func(stats search.ScanStats) {
			if stats.FilesScanned%50 == 0 {
				elapsed := time.Since(startTime)
				speed := float64(stats.FilesScanned) / elapsed.Seconds()
				writer.WriteString(fmt.Sprintf("\r📊 进度: 已扫描 %d 个文件，匹配 %d 个，速度 %.0f 文件/秒",
					stats.FilesScanned, stats.FilesMatched, speed))
				writer.Flush()
			}
		},
	}

	writer.WriteString("\n🔍 正在搜索...\n")
	writer.Flush()

	// 收集所有结果
	results := search.Search(config)

	// 清除最后的进度行
	writer.WriteString("\r" + strings.Repeat(" ", 80) + "\r")

	elapsed := time.Since(startTime)
	totalCount := len(results)

	if totalCount == 0 {
		writer.WriteString("未找到匹配的结果\n")
		writer.Flush()
		pause(writer)
		return
	}

	writer.WriteString(fmt.Sprintf("✅ 找到 %d 个匹配结果，耗时 %v\n", totalCount, elapsed.Round(time.Millisecond)))
	writer.Flush()

	if usePaged {
		displayPagedResults(writer, results, combinedKeyword, pageSize)
	} else {
		search.PrintResults(results, combinedKeyword)
	}

	writer.Flush()
	pause(writer)
}

