package main

import (
	"flag"
	"fmt"
	"os"
	"searching-tool/internal/searcher"
)

func main() {
	// 命令行参数解析
	var searchFile bool
	var workerNum int
	var bufSize int

	flag.BoolVar(&searchFile, "v", false, "搜索文件名（而非文件内容）")
	flag.IntVar(&workerNum, "w", 4, "工作协程数量")
	flag.IntVar(&bufSize, "b", 100, "通道缓冲区大小")
	flag.Usage = func() {
		fmt.Println("使用方式：./catch [-v] [-w 协程数] [-b 缓冲区大小] [目录] [关键词]")
		fmt.Println("示例:")
		fmt.Println("  ./catch ./src func              # 搜索 ./src 目录下包含 'func' 的文件内容")
		fmt.Println("  ./catch -v ./src .go            # 搜索 ./src 目录下文件名包含 '.go' 的文件")
		fmt.Println("  ./catch -w 8 -b 200 ./src main  # 使用 8 个协程，缓冲区 200 进行搜索")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("错误：参数不足")
		fmt.Println("正确格式：./catch [-v] [-w 协程数] [-b 缓冲区大小] [目录] [关键词]")
		os.Exit(1)
	}

	dir := args[0]
	keyword := args[1]

	// 验证目录是否存在
	info, err := os.Stat(dir)
	if err != nil {
		fmt.Printf("错误：无法访问目录 '%s': %v\n", dir, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Printf("错误：'%s' 不是目录\n", dir)
		os.Exit(1)
	}

	// 创建配置
	cfg := searcher.Config{
		WorkerNum:      workerNum,
		ChanBufferSize: bufSize,
	}

	// 创建并运行搜索引擎
	engine := searcher.NewSearchEngine(dir, keyword, searchFile, cfg)
	engine.Run()

	fmt.Println("\n搜索结束")
}
