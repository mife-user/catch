package searcher

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

// Worker 工作协程 - 负责执行单个搜索任务
type Worker struct {
	id    int
	stats *Stats
}

// NewWorker 创建 worker
func NewWorker(id int, stats *Stats) *Worker {
	return &Worker{
		id:    id,
		stats: stats,
	}
}

// Process 处理单个搜索任务
func (w *Worker) Process(task SearchTask) []SearchResult {
	var results []SearchResult

	if task.SearchType {
		// 文件名搜索模式
		if strings.Contains(task.FilePath, task.KeyWord) {
			atomic.AddInt64(&w.stats.FilesScanned, 1)
			atomic.AddInt64(&w.stats.MatchesFound, 1)
			results = append(results, SearchResult{
				FilePath: task.FilePath,
				LinNum:   0,
				Content:  "Matches Filename",
			})
		}
		return results
	}

	// 文件内容搜索模式
	file, err := os.Open(task.FilePath)
	if err != nil {
		if os.IsPermission(err) {
			// 权限不足，跳过
			atomic.AddInt64(&w.stats.FilesSkipped, 1)
			atomic.AddInt64(&w.stats.ErrorsOccurred, 1)
		} else {
			atomic.AddInt64(&w.stats.ErrorsOccurred, 1)
		}
		return nil
	}
	defer file.Close()

	atomic.AddInt64(&w.stats.FilesScanned, 1)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		atomic.AddInt64(&w.stats.LinesScanned, 1)
		text := scanner.Text()
		if strings.Contains(text, task.KeyWord) {
			atomic.AddInt64(&w.stats.MatchesFound, 1)
			results = append(results, SearchResult{
				FilePath: task.FilePath,
				LinNum:   lineNum,
				Content:  text,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		atomic.AddInt64(&w.stats.ErrorsOccurred, 1)
	}

	return results
}

// WorkerPool 工作池 - 管理多个 worker 协程
type WorkerPool struct {
	workerNum  int
	taskChan   chan SearchTask
	resultChan chan SearchResult
	stats      *Stats
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWorkerPool 创建工作池
func NewWorkerPool(cfg Config) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workerNum:  cfg.WorkerNum,
		taskChan:   make(chan SearchTask, cfg.ChanBufferSize),
		resultChan: make(chan SearchResult, cfg.ChanBufferSize),
		stats:      &Stats{},
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerNum; i++ {
		wp.wg.Add(1)
		go func(id int) {
			defer wp.wg.Done()
			worker := NewWorker(id, wp.stats)
			for task := range wp.taskChan {
				results := worker.Process(task)
				for _, result := range results {
					select {
					case <-wp.ctx.Done():
						return
					case wp.resultChan <- result:
					}
				}
			}
		}(i)
	}
}

// Submit 提交任务
func (wp *WorkerPool) Submit(task SearchTask) {
	select {
	case <-wp.ctx.Done():
	case wp.taskChan <- task:
	}
}

// TaskChan 获取任务输入通道
func (wp *WorkerPool) TaskChan() chan<- SearchTask {
	return wp.taskChan
}

// CloseTasks 关闭任务输入通道，通知 worker 结束
func (wp *WorkerPool) CloseTasks() {
	close(wp.taskChan)
}

// Wait 等待所有 worker 完成
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

// Results 获取结果通道
func (wp *WorkerPool) Results() <-chan SearchResult {
	return wp.resultChan
}

// Close 关闭工作池
func (wp *WorkerPool) Close() {
	wp.cancel()
}

// Stats 获取统计信息
func (wp *WorkerPool) Stats() Stats {
	return *wp.stats
}
