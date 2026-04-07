package cli

import (
	"fmt"
	"math"
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
	width      int
	visible    bool
}

// NewProgressBar 创建新的进度条
func NewProgressBar(total int, width int) *ProgressBar {
	return &ProgressBar{
		total:      total,
		current:    0,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
		width:      width,
		visible:    true,
	}
}

// Update 更新进度
func (pb *ProgressBar) Update(current int, stats search.ScanStats) {
	pb.current = current
	now := time.Now()

	// 限制更新频率（每 100ms 更新一次）
	if now.Sub(pb.lastUpdate) < 100*time.Millisecond {
		return
	}
	pb.lastUpdate = now

	pb.render(stats)
}

// Finish 完成进度条
func (pb *ProgressBar) Finish(stats search.ScanStats) {
	pb.current = pb.total
	pb.render(stats)
	fmt.Println() // 换行
}

// render 渲染进度条
func (pb *ProgressBar) render(stats search.ScanStats) {
	if !pb.visible {
		return
	}

	// 计算进度百分比
	var percent float64
	if pb.total > 0 {
		percent = float64(pb.current) / float64(pb.total) * 100
	} else {
		percent = 0
	}

	// 计算进度条
	filledWidth := int(math.Round(percent / 100.0 * float64(pb.width)))
	if filledWidth > pb.width {
		filledWidth = pb.width
	}
	emptyWidth := pb.width - filledWidth

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", emptyWidth)

	// 计算速度
	elapsed := time.Since(pb.startTime).Seconds()
	var speed float64
	if elapsed > 0 {
		speed = float64(stats.FilesScanned) / elapsed
	}

	// 计算预计剩余时间
	var eta string
	if speed > 0 && pb.total > 0 {
		remaining := float64(pb.total-pb.current) / speed
		if remaining < 60 {
			eta = fmt.Sprintf("%.0fs", remaining)
		} else {
			eta = fmt.Sprintf("%.0fm", remaining/60)
		}
	} else {
		eta = "--"
	}

	// 渲染进度条
	fmt.Printf("\r🔍 [%s%s] %5.1f%% | 已扫描: %d/%d | 匹配: %d | 速度: %.0f 文件/秒 | 预计: %s",
		filled, empty, percent, stats.FilesScanned, pb.total, stats.FilesMatched, speed, eta)
}

// Hide 隐藏进度条
func (pb *ProgressBar) Hide() {
	pb.visible = false
}

// Show 显示进度条
func (pb *ProgressBar) Show() {
	pb.visible = true
}
