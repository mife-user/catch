package cli

import (
	"bufio"
	"catch/internal/config"
	"catch/internal/history"
	"catch/internal/search"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// searchHistory 搜索历史功能
func searchHistory(reader *bufio.Reader, writer *SimpleWriter, historyMgr *history.HistoryManager, cfg *config.Config) {
	if historyMgr == nil {
		writer.WriteString("❌ 历史记录管理器未初始化\n")
		writer.Flush()
		pause(writer)
		return
	}

	if historyMgr.Count() == 0 {
		writer.WriteString("📜 暂无搜索历史\n")
		writer.Flush()
		pause(writer)
		return
	}

	// 显示历史记录
	entries := historyMgr.GetRecentEntries(20) // 最多显示 20 条
	writer.WriteString("📜 最近搜索历史:\n\n")

	for i, entry := range entries {
		writer.WriteString(fmt.Sprintf("  [%d] %s - %s (路径: %s)\n",
			entry.ID, entry.FormatTimestamp(), entry.GetDescription(), entry.Path))
		if i < len(entries)-1 {
			writer.WriteString("\n")
		}
	}

	writer.WriteString("\n" + strings.Repeat("─", 50) + "\n")
	writer.WriteString("操作: [编号] 重新搜索, [c] 清空历史, [q] 返回: ")
	writer.Flush()

	input := readLine(reader)
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "q" || input == "quit" || input == "exit" {
		return
	}

	if input == "c" || input == "clear" {
		writer.WriteString("确认清空所有历史记录？(y/n): ")
		writer.Flush()

		confirm := readLine(reader)
		if strings.TrimSpace(strings.ToLower(confirm)) == "y" {
			historyMgr.Clear()
			writer.WriteString("✅ 历史记录已清空\n")
			writer.Flush()
		}
		return
	}

	// 尝试解析编号
	id, err := strconv.Atoi(input)
	if err != nil {
		writer.WriteString("❌ 无效输入\n")
		writer.Flush()
		return
	}

	// 获取历史记录
	entry, err := historyMgr.GetEntry(id)
	if err != nil {
		writer.WriteString(fmt.Sprintf("❌ %v\n", err))
		writer.Flush()
		return
	}

	// 重新执行搜索
	writer.WriteString(fmt.Sprintf("\n🔍 重新搜索: %s\n", entry.GetDescription()))
	writer.Flush()

	// 根据搜索模式执行不同的搜索
	if entry.UseRegex {
		// 正则表达式搜索
		regexSearchWithEntry(reader, writer, entry, cfg)
	} else if entry.SearchMode == "multi_and" || entry.SearchMode == "multi_or" {
		// 多关键字搜索
		multiKeywordSearchWithEntry(reader, writer, entry, cfg)
	} else {
		// 普通搜索
		searchContentWithEntry(reader, writer, entry, cfg)
	}
}

// configManager 配置管理功能
func configManager(reader *bufio.Reader, writer *SimpleWriter, cfg *config.Config) {
	writer.WriteString("⚙️  配置管理\n\n")
	writer.WriteString("当前配置:\n")
	writer.WriteString(fmt.Sprintf("  默认递归: %v\n", cfg.DefaultRecursive))
	writer.WriteString(fmt.Sprintf("  默认每页大小: %d\n", cfg.DefaultPageSize))
	writer.WriteString(fmt.Sprintf("  默认上下文行数: %d\n", cfg.DefaultContextLines))
	writer.WriteString(fmt.Sprintf("  默认最大协程数: %d\n", cfg.DefaultMaxGoroutine))
	writer.WriteString(fmt.Sprintf("  默认最大文件大小: %d MB\n", cfg.DefaultMaxFileSize/1024/1024))
	writer.WriteString(fmt.Sprintf("  默认最大匹配数: %d\n", cfg.DefaultMaxMatches))
	writer.WriteString(fmt.Sprintf("  默认导出格式: %s\n", cfg.DefaultExportFormat))
	writer.WriteString(fmt.Sprintf("  历史记录最大条数: %d\n", cfg.HistoryMaxEntries))
	writer.WriteString(fmt.Sprintf("  跳过目录: %v\n", cfg.SkipDirs))
	writer.WriteString(fmt.Sprintf("  跳过扩展名: %v\n", cfg.SkipExtensions))
	writer.WriteString(fmt.Sprintf("  主题: %s\n", cfg.Theme))
	writer.WriteString("\n" + strings.Repeat("─", 50) + "\n")
	writer.WriteString("操作: [1] 修改配置, [2] 生成配置文件, [q] 返回: ")
	writer.Flush()

	input := readLine(reader)
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "q" {
		return
	}

	switch input {
	case "1":
		modifyConfig(reader, writer, cfg)
	case "2":
		generateConfigFile(reader, writer, cfg)
	default:
		writer.WriteString("❌ 无效输入\n")
		writer.Flush()
	}
}

// modifyConfig 修改配置
func modifyConfig(reader *bufio.Reader, writer *SimpleWriter, cfg *config.Config) {
	writer.WriteString("\n修改配置:\n")
	writer.WriteString("  [1] 默认递归\n")
	writer.WriteString("  [2] 默认每页大小\n")
	writer.WriteString("  [3] 默认上下文行数\n")
	writer.WriteString("  [4] 默认最大协程数\n")
	writer.WriteString("  [5] 历史记录最大条数\n")
	writer.WriteString("  [q] 返回\n\n")
	writer.WriteString("请选择: ")
	writer.Flush()

	input := readLine(reader)
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		writer.WriteString("是否默认递归？(y/n): ")
		writer.Flush()
		val := readLine(reader)
		cfg.DefaultRecursive = strings.TrimSpace(strings.ToLower(val)) == "y"
		writer.WriteString("✅ 已更新\n")
	case "2", "3", "4", "5":
		writer.WriteString("请输入新值: ")
		writer.Flush()
		val := readLine(reader)
		numVal, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil || numVal <= 0 {
			writer.WriteString("❌ 无效值\n")
			writer.Flush()
			return
		}
		switch input {
		case "2":
			cfg.DefaultPageSize = numVal
		case "3":
			cfg.DefaultContextLines = numVal
		case "4":
			cfg.DefaultMaxGoroutine = numVal
		case "5":
			cfg.HistoryMaxEntries = numVal
		}
		writer.WriteString("✅ 已更新\n")
	case "q":
		return
	default:
		writer.WriteString("❌ 无效输入\n")
		writer.Flush()
		return
	}

	writer.Flush()
	pause(writer)
}

// generateConfigFile 生成配置文件
func generateConfigFile(reader *bufio.Reader, writer *SimpleWriter, cfg *config.Config) {
	writer.WriteString("\n生成配置文件:\n")
	writer.WriteString("  [1] 当前目录 (.catchrc)\n")
	writer.WriteString("  [2] 用户主目录 (~/.catchrc)\n")
	writer.WriteString("  [q] 返回\n\n")
	writer.WriteString("请选择: ")
	writer.Flush()

	input := readLine(reader)
	input = strings.TrimSpace(input)

	var configPath string
	switch input {
	case "1":
		configPath = ".catchrc"
	case "2":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			writer.WriteString("❌ 获取用户主目录失败\n")
			writer.Flush()
			return
		}
		configPath = filepath.Join(homeDir, ".catchrc")
	case "q":
		return
	default:
		writer.WriteString("❌ 无效输入\n")
		writer.Flush()
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); err == nil {
		writer.WriteString("文件已存在，是否覆盖？(y/n): ")
		writer.Flush()
		confirm := readLine(reader)
		if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
			return
		}
	}

	if err := config.SaveConfig(cfg, configPath); err != nil {
		writer.WriteString(fmt.Sprintf("❌ 生成配置文件失败: %v\n", err))
		writer.Flush()
		return
	}

	writer.WriteString(fmt.Sprintf("✅ 配置文件已生成: %s\n", configPath))
	writer.Flush()
	pause(writer)
}

// searchContentWithEntry 从历史记录重新搜索
func searchContentWithEntry(reader *bufio.Reader, writer *SimpleWriter, entry *history.HistoryEntry, cfg *config.Config) {
	if entry.Keyword == "" {
		writer.WriteString("❌ 关键字不能为空\n")
		writer.Flush()
		pause(writer)
		return
	}

	searchPath := entry.Path
	if searchPath == "" {
		searchPath = "."
	}

	if _, err := os.Stat(searchPath); err != nil {
		writer.WriteString(fmt.Sprintf("❌ 路径不存在：%s\n", searchPath))
		writer.Flush()
		pause(writer)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	startTime := time.Now()

	searchConfig := search.SearchConfig{
		Keyword:      entry.Keyword,
		Path:         searchPath,
		Recursive:    entry.Recursive,
		MaxGoroutine: cfg.DefaultMaxGoroutine,
		ContextLines: entry.ContextLines,
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

	results := search.Search(searchConfig)

	writer.WriteString("\r" + strings.Repeat(" ", 80) + "\r")

	elapsed := time.Since(startTime)
	totalCount := len(results)

	if totalCount == 0 {
		writer.WriteString("未找到匹配的结果\n")
	} else {
		writer.WriteString(fmt.Sprintf("✅ 找到 %d 个匹配结果，耗时 %v\n", totalCount, elapsed.Round(time.Millisecond)))
		search.PrintResults(results, entry.Keyword)
	}
	writer.Flush()
	pause(writer)
}

// regexSearchWithEntry 从历史记录重新正则搜索
func regexSearchWithEntry(reader *bufio.Reader, writer *SimpleWriter, entry *history.HistoryEntry, cfg *config.Config) {
	if entry.RegexPattern == "" {
		writer.WriteString("❌ 正则表达式不能为空\n")
		writer.Flush()
		pause(writer)
		return
	}

	searchPath := entry.Path
	if searchPath == "" {
		searchPath = "."
	}

	if _, err := os.Stat(searchPath); err != nil {
		writer.WriteString(fmt.Sprintf("❌ 路径不存在：%s\n", searchPath))
		writer.Flush()
		pause(writer)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	startTime := time.Now()

	searchConfig := search.SearchConfig{
		RegexPattern: nil,
		UseRegex:     true,
		Path:         searchPath,
		Recursive:    entry.Recursive,
		MaxGoroutine: cfg.DefaultMaxGoroutine,
		ContextLines: entry.ContextLines,
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

	writer.WriteString(fmt.Sprintf("\n🔍 正在使用正则搜索: %s\n", entry.RegexPattern))
	writer.Flush()

	results := search.Search(searchConfig)

	writer.WriteString("\r" + strings.Repeat(" ", 80) + "\r")

	elapsed := time.Since(startTime)
	totalCount := len(results)

	if totalCount == 0 {
		writer.WriteString("未找到匹配的结果\n")
	} else {
		writer.WriteString(fmt.Sprintf("✅ 找到 %d 个匹配结果，耗时 %v\n", totalCount, elapsed.Round(time.Millisecond)))
		search.PrintResults(results, entry.RegexPattern)
	}
	writer.Flush()
	pause(writer)
}

// multiKeywordSearchWithEntry 从历史记录重新多关键字搜索
func multiKeywordSearchWithEntry(reader *bufio.Reader, writer *SimpleWriter, entry *history.HistoryEntry, cfg *config.Config) {
	if len(entry.Keywords) == 0 {
		writer.WriteString("❌ 关键字不能为空\n")
		writer.Flush()
		pause(writer)
		return
	}

	searchPath := entry.Path
	if searchPath == "" {
		searchPath = "."
	}

	if _, err := os.Stat(searchPath); err != nil {
		writer.WriteString(fmt.Sprintf("❌ 路径不存在：%s\n", searchPath))
		writer.Flush()
		pause(writer)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	startTime := time.Now()

	searchConfig := search.SearchConfig{
		Keywords:     entry.Keywords,
		SearchMode:   entry.SearchMode,
		Path:         searchPath,
		Recursive:    entry.Recursive,
		MaxGoroutine: cfg.DefaultMaxGoroutine,
		ContextLines: entry.ContextLines,
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

	modeDesc := "AND"
	if entry.SearchMode == "multi_or" {
		modeDesc = "OR"
	}
	combinedKeyword := strings.Join(entry.Keywords, fmt.Sprintf(" [%s] ", modeDesc))

	writer.WriteString(fmt.Sprintf("\n🔍 正在多关键字搜索 [%s]: %s\n", modeDesc, combinedKeyword))
	writer.Flush()

	results := search.Search(searchConfig)

	writer.WriteString("\r" + strings.Repeat(" ", 80) + "\r")

	elapsed := time.Since(startTime)
	totalCount := len(results)

	if totalCount == 0 {
		writer.WriteString("未找到匹配的结果\n")
	} else {
		writer.WriteString(fmt.Sprintf("✅ 找到 %d 个匹配结果，耗时 %v\n", totalCount, elapsed.Round(time.Millisecond)))
		search.PrintResults(results, combinedKeyword)
	}
	writer.Flush()
	pause(writer)
}
