//go:build windows
// +build windows

package cli

import (
	"syscall"
)

func init() {
	// 设置控制台输出代码页为 UTF-8 (65001)
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	setConsoleOutputCP.Call(65001)
}
