package searcher

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Collector 结果收集器 - 负责收集和输出结果
type Collector struct {
	count int64
}

// NewCollector 创建结果收集器
func NewCollector() *Collector {
	return &Collector{}
}

// Collect 收集并输出单个结果
func (c *Collector) Collect(result SearchResult) {
	num := atomic.AddInt64(&c.count, 1)
	fmt.Printf("[%d] 路径：%s\n    行数：%d\n    内容：%s\n\n", num, result.FilePath, result.LinNum, result.Content)
}

// Count 获取已收集的结果数量
func (c *Collector) Count() int64 {
	return atomic.LoadInt64(&c.count)
}

// SilentCollector 静默收集器 - 不输出结果，仅统计
type SilentCollector struct {
	mu    sync.Mutex
	items []SearchResult
}

// NewSilentCollector 创建静默收集器
func NewSilentCollector() *SilentCollector {
	return &SilentCollector{
		items: make([]SearchResult, 0),
	}
}

// Collect 收集结果但不输出
func (sc *SilentCollector) Collect(result SearchResult) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.items = append(sc.items, result)
}

// Results 获取所有收集的结果
func (sc *SilentCollector) Results() []SearchResult {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.items
}

// Count 获取已收集的结果数量
func (sc *SilentCollector) Count() int64 {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return int64(len(sc.items))
}
