package main

import (
	"catch/internal/cli"
	"catch/internal/search"
	"context"
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		// 无参数时进入交互模式
		cli.RunInteractive()
		return
	}

	// 命令行模式
	command := os.Args[1]
	switch command {
	case "search", "s":
		runSearch(os.Args[2:])
	case "add-path":
		cli.AddToPath(&cli.SimpleWriter{Writer: os.Stdout})
	case "help", "h":
		printHelp()
	default:
		fmt.Printf("未知命令：%s\n", command)
		printHelp()
	}
}

func runSearch(args []string) {
	if len(args) < 1 {
		fmt.Println("用法：catch search <关键字> [选项]")
		fmt.Println("选项:")
		fmt.Println("  -r, --recursive    递归搜索子目录")
		fmt.Println("  -t, --type         文件类型过滤 (如：.go,.txt)")
		fmt.Println("  -p, --path         搜索路径 (默认为当前目录)")
		return
	}

	keyword := args[0]
	recursive := false
	fileType := ""
	searchPath := "."

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-r", "--recursive":
			recursive = true
		case "-t", "--type":
			if i+1 < len(args) {
				i++
				fileType = args[i]
			}
		case "-p", "--path":
			if i+1 < len(args) {
				i++
				searchPath = args[i]
			}
		}
	}

	// 创建带超时的 context（120秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	config := search.SearchConfig{
		Keyword:      keyword,
		Path:         searchPath,
		Recursive:    recursive,
		FileType:     fileType,
		MaxGoroutine: 10,
		Context:      ctx,
	}

	results := search.Search(config)
	search.PrintResults(results, keyword)
}

func printHelp() {
	fmt.Println("Catch - 本地文件夹内容搜索工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  catch                    进入交互模式")
	fmt.Println("  catch search <关键字>    搜索文件内容和文件名")
	fmt.Println("  catch add-path           添加到系统环境变量")
	fmt.Println("  catch help               显示帮助信息")
	fmt.Println()
	fmt.Println("搜索选项:")
	fmt.Println("  -r, --recursive    递归搜索子目录")
	fmt.Println("  -t, --type         文件类型过滤 (如：.go,.txt)")
	fmt.Println("  -p, --path         搜索路径 (默认为当前目录)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  catch search hello -r")
	fmt.Println("  catch search func -t .go -p ./src")
}
