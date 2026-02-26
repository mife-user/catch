package searcher

import (
	"context"
	"fmt"
)

// SearchEngine 搜索引擎 - 统一入口，协调各组件
type SearchEngine struct {
	config    Config
	scanner   *DirScanner
	pool      *WorkerPool
	collector *Collector
}

// NewSearchEngine 创建搜索引擎
func NewSearchEngine(rootPath, keyword string, searchType bool, cfg Config) *SearchEngine {
	return &SearchEngine{
		config:    cfg,
		scanner:   NewDirScanner(rootPath, keyword, searchType),
		pool:      NewWorkerPool(cfg),
		collector: NewCollector(),
	}
}

// Run 执行搜索
func (se *SearchEngine) Run() Stats {
	modeStr := map[bool]string{true: "文件名", false: "文件内容"}[se.scanner.searchType]
	fmt.Printf("开始搜索 - 目录：%s, 关键词：%s, 模式：%s\n",
		se.scanner.rootPath,
		se.scanner.keyword,
		modeStr)
	fmt.Printf("工作协程数：%d, 缓冲区大小：%d\n\n", se.config.WorkerNum, se.config.ChanBufferSize)

	// 启动工作池
	se.pool.Start()

	// 启动结果收集协程
	collectDone := make(chan struct{})
	go func() {
		for result := range se.pool.Results() {
			se.collector.Collect(result)
		}
		close(collectDone)
	}()

	// 扫描目录并提交任务
	scanErr := se.scanner.Scan(context.Background(), se.pool.TaskChan(), se.pool.stats)
	if scanErr != nil {
		fmt.Printf("扫描目录出错：%v\n", scanErr)
	}

	// 关闭任务通道，等待 worker 完成
	se.pool.CloseTasks()
	se.pool.Wait()

	// 关闭结果通道，等待收集完成
	close(se.pool.resultChan)
	<-collectDone

	// 关闭工作池
	se.pool.Close()

	stats := se.pool.Stats()
	se.printStats(stats)

	return stats
}

// RunSilent 执行搜索但不输出结果 (返回结果切片)
func (se *SearchEngine) RunSilent() ([]SearchResult, Stats) {
	// 启动工作池
	se.pool.Start()

	// 使用静默收集器
	silentCollector := NewSilentCollector()
	collectDone := make(chan struct{})
	go func() {
		for result := range se.pool.Results() {
			silentCollector.Collect(result)
		}
		close(collectDone)
	}()

	// 扫描目录并提交任务
	scanErr := se.scanner.Scan(context.Background(), se.pool.TaskChan(), se.pool.stats)
	if scanErr != nil {
		fmt.Printf("扫描目录出错：%v\n", scanErr)
	}

	// 关闭任务通道，等待 worker 完成
	se.pool.CloseTasks()
	se.pool.Wait()

	// 关闭结果通道，等待收集完成
	close(se.pool.resultChan)
	<-collectDone

	// 关闭工作池
	se.pool.Close()

	stats := se.pool.Stats()
	return silentCollector.Results(), stats
}

func (se *SearchEngine) printStats(stats Stats) {
	fmt.Printf("\n========== 搜索完成 ==========\n")
	fmt.Printf("扫描文件数：%d\n", stats.FilesScanned)
	fmt.Printf("扫描行数：%d\n", stats.LinesScanned)
	fmt.Printf("匹配结果：%d\n", stats.MatchesFound)
	fmt.Printf("发生错误：%d\n", stats.ErrorsOccurred)
	fmt.Printf("跳过文件 (权限不足): %d\n", stats.FilesSkipped)
	fmt.Printf("==============================\n")
}
