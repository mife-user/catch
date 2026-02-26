package searcher

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
)

// DirScanner 目录扫描器 - 负责遍历目录生成任务
type DirScanner struct {
	rootPath   string
	keyword    string
	searchType bool
}

// NewDirScanner 创建目录扫描器
func NewDirScanner(rootPath, keyword string, searchType bool) *DirScanner {
	return &DirScanner{
		rootPath:   rootPath,
		keyword:    keyword,
		searchType: searchType,
	}
}

// Scan 扫描目录，将任务发送到任务通道
// 如果遇到权限不足的目录或文件，会跳过并记录到 stats 中
func (ds *DirScanner) Scan(ctx context.Context, taskChan chan<- SearchTask, stats *Stats) error {
	err := filepath.WalkDir(ds.rootPath, func(path string, info fs.DirEntry, err error) error {
		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 处理访问错误 (包括权限不足)
		if err != nil {
			if os.IsPermission(err) {
				// 权限不足，跳过并记录
				stats.ErrorsOccurred++
				stats.FilesSkipped++
				return nil // 返回 nil 继续遍历其他文件
			}
			return err // 其他错误返回
		}

		// 检查是否有权限访问此文件/目录
		if !ds.canAccess(path) {
			stats.ErrorsOccurred++
			stats.FilesSkipped++
			// 如果是目录，返回 filepath.SkipDir 跳过整个目录
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil // 文件则跳过继续
		}

		// 如果是目录且不是文件名搜索模式，跳过
		if info.IsDir() && !ds.searchType {
			return nil
		}

		task := SearchTask{
			FilePath:   path,
			KeyWord:    ds.keyword,
			SearchType: ds.searchType,
		}

		// 发送任务
		select {
		case <-ctx.Done():
			return ctx.Err()
		case taskChan <- task:
		}

		return nil
	})

	return err
}

// canAccess 检查是否有权限访问文件/目录
func (ds *DirScanner) canAccess(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return !os.IsPermission(err)
	}
	return true
}
