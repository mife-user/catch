package cli

import (
	"fmt"
	"strings"
	"time"

	"catch/internal/search"
)

// ProgressBar 进度条
type ProgressBar struct {
	total      int
	current    int
	startTime  time.Time
	lastUpdate time.Time
	lastCount  int
	width      int
	visible    bool
	writer     *SimpleWriter
}

// NewProgressBar 创建新的进度条
func NewProgressBar(writer *SimpleWriter, total int, width int) *ProgressBar {
	return &ProgressBar{
		total:      total,
		current:    0,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
		lastCount:  0,
		width:      width,
		visible:    true,
		writer:     writer,
	}
}

// Update 更新进度
func (pb *ProgressBar) Update(current int, stats search.ScanStats) {
	if !pb.visible {
		return
	}

	pb.current = current
	now := time.Now()

	// 限制更新频率（每 200ms 更新一次）
	if now.Sub(pb.lastUpdate) < 200*time.Millisecond && stats.FilesScanned == pb.lastCount {
		return
	}
	pb.lastUpdate = now
	pb.lastCount = stats.FilesScanned

	pb.render(stats)
}

// Finish 完成进度条
func (pb *ProgressBar) Finish(stats search.ScanStats) {
	if !pb.visible {
		return
	}
	pb.render(stats)
	pb.writer.WriteString("\n")
	pb.writer.Flush()
}

// render 渲染进度条
func (pb *ProgressBar) render(stats search.ScanStats) {
	// 计算进度百分比
	var percent float64
	showPercent := pb.total > 0

	if showPercent {
		percent = float64(pb.current) / float64(pb.total) * 100
		if percent > 100 {
			percent = 100
		}
	}

	// 计算速度
	elapsed := time.Since(pb.startTime).Seconds()
	var speed float64
	if elapsed > 0 {
		speed = float64(stats.FilesScanned) / elapsed
	}

	// 渲染进度条
	var progressBarStr string
	if showPercent {
		// 显示进度条模式
		filledWidth := int(float64(pb.width) * percent / 100.0)
		if filledWidth > pb.width {
			filledWidth = pb.width
		}
		emptyWidth := pb.width - filledWidth

		filled := strings.Repeat("█", filledWidth)
		empty := strings.Repeat("░", emptyWidth)

		progressBarStr = fmt.Sprintf("\r🔍 [%s%s] %5.1f%%", filled, empty, percent)
	} else {
		// 旋转动画模式（未知总数）
		spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		idx := int(elapsed*5) % len(spinners)
		progressBarStr = fmt.Sprintf("\r🔍 %s 搜索中...", spinners[idx])
	}

	// 输出进度信息
	output := fmt.Sprintf("%s | 已扫描: %d | 匹配: %d | 速度: %.0f 文件/秒",
		progressBarStr, stats.FilesScanned, stats.FilesMatched, speed)

	// 清除整行并输出
	pb.writer.WriteString("\r" + strings.Repeat(" ", 100) + output)
	pb.writer.Flush()
}

// Hide 隐藏进度条
func (pb *ProgressBar) Hide() {
	pb.visible = false
}

// Show 显示进度条
func (pb *ProgressBar) Show() {
	pb.visible = true
}

// ClearLine 清除当前行
func ClearLine() {
	fmt.Print("\r" + strings.Repeat(" ", 100) + "\r")
}

// IsTerminal 检查是否是终端
func IsTerminal() bool {
	return true
}
