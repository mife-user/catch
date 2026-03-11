package cli

import (
	"bufio"
	"catch/internal/search"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

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

// init 初始化 Windows 控制台编码
func init() {
	if runtime.GOOS == "windows" {
		// 设置控制台输出代码页为 UTF-8 (65001)
		kernel32 := syscall.NewLazyDLL("kernel32.dll")
		setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
		setConsoleOutputCP.Call(65001)
	}
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
	menuItems := []MenuItem{
		{Title: "🔍 搜索文件内容", Action: func() { searchContent(reader, writer) }, Shortcut: "1"},
		{Title: "📁 搜索文件名", Action: func() { searchFilename(reader, writer) }, Shortcut: "2"},
		{Title: "⚙️  高级搜索", Action: func() { advancedSearch(reader, writer) }, Shortcut: "3"},
		{Title: "➕ 添加到环境变量", Action: func() { AddToPath(writer) }, Shortcut: "4"},
		{Title: "❌ 退出", Action: func() {}, Shortcut: "q"},
	}

	for {
		printMenu(writer, menuItems)
		writer.WriteString("\n请选择功能 (1-4 或 q): ")
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
			writer.WriteString("❌ 无效选择，请重试\n")
			writer.Flush()
			continue
		}

		menuItems[selected-1].Action()
	}
}

func runSimpleUI(reader *bufio.Reader, writer *SimpleWriter) {
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
			searchContent(reader, writer)
		case "2":
			searchFilename(reader, writer)
		case "3":
			advancedSearch(reader, writer)
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
	// 清屏
	clearScreen()

	writer.WriteString("╔════════════════════════════════════════════╗\n")
	writer.WriteString("║     🎯 Catch - 文件搜索工具                ║\n")
	writer.WriteString("╠════════════════════════════════════════════╣\n")

	for _, item := range items {
		// 计算显示宽度：中文和 Emoji 算 2 个字符，英文算 1 个
		title := item.Title
		displayWidth := getDisplayWidth(title)
		padding := 38 - displayWidth
		if padding < 0 {
			padding = 0
		}
		writer.WriteString(fmt.Sprintf("║  [%s] %s%s ║\n", item.Shortcut, title, strings.Repeat(" ", padding)))
	}

	writer.WriteString("╚════════════════════════════════════════════╝\n")
	writer.Flush()
}

// getDisplayWidth 计算字符串的显示宽度（中文和 Emoji 算 2 个字符）
func getDisplayWidth(s string) int {
	width := 0
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			// 中文字符
			width += 2
		} else if r >= 0x1F000 {
			// Emoji 字符（基本在 U+1F000 以上）
			width += 2
		} else if r >= 0x3000 && r <= 0x303F {
			// CJK 符号和标点
			width += 2
		} else if r >= 0xFF00 && r <= 0xFFEF {
			// 全角字符
			width += 2
		} else {
			// ASCII 字符
			width += 1
		}
	}
	return width
}

func clearScreen() {
	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/c", "cls").Run()
	} else {
		exec.Command("clear").Run()
	}
}

func searchContent(reader *bufio.Reader, writer *SimpleWriter) {
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

	writer.WriteString("是否递归搜索子目录？(y/n): ")
	writer.Flush()

	recInput := readLine(reader)
	recursive := strings.TrimSpace(strings.ToLower(recInput)) == "y"

	config := search.SearchConfig{
		Keyword:      keyword,
		Path:         ".",
		Recursive:    recursive,
		MaxGoroutine: 10,
	}

	writer.WriteString("\n🔍 正在搜索...\n")
	writer.Flush()

	results := search.Search(config)
	search.PrintResults(results, keyword)
	pause(writer)
}

func searchFilename(reader *bufio.Reader, writer *SimpleWriter) {
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

	writer.WriteString("是否递归搜索子目录？(y/n): ")
	writer.Flush()

	recInput := readLine(reader)
	recursive := strings.TrimSpace(strings.ToLower(recInput)) == "y"

	config := search.SearchConfig{
		Keyword:      keyword,
		Path:         ".",
		Recursive:    recursive,
		MaxGoroutine: 10,
	}

	writer.WriteString("\n🔍 正在搜索...\n")
	writer.Flush()

	results := search.Search(config)

	// 只显示文件名匹配的结果
	filenameResults := make([]search.SearchResult, 0)
	for _, r := range results {
		if r.MatchType == "filename" {
			filenameResults = append(filenameResults, r)
		}
	}

	search.PrintResults(filenameResults, keyword)
	pause(writer)
}

func advancedSearch(reader *bufio.Reader, writer *SimpleWriter) {
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

	writer.WriteString("是否递归搜索子目录？(y/n): ")
	writer.Flush()

	recInput := readLine(reader)
	recursive := strings.TrimSpace(strings.ToLower(recInput)) == "y"

	writer.WriteString("文件类型过滤 (如：.go,.txt，留空表示全部): ")
	writer.Flush()

	fileType := readLine(reader)
	fileType = strings.TrimSpace(fileType)

	writer.WriteString("并发协程数 (默认 10): ")
	writer.Flush()

	goroutineInput := readLine(reader)
	goroutines, err := strconv.Atoi(strings.TrimSpace(goroutineInput))
	if err != nil || goroutines <= 0 {
		goroutines = 10
	}

	config := search.SearchConfig{
		Keyword:      keyword,
		Path:         path,
		Recursive:    recursive,
		FileType:     fileType,
		MaxGoroutine: goroutines,
	}

	writer.WriteString("\n🔍 正在搜索...\n")
	writer.Flush()

	results := search.Search(config)
	search.PrintResults(results, keyword)
	pause(writer)
}

func pause(writer *SimpleWriter) {
	writer.WriteString("\n按 Enter 键继续...")
	writer.Flush()
	readLine(bufio.NewReader(os.Stdin))
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
