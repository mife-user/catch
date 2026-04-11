package service

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func moveToSystemTrash(filePath string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		escaped := strings.ReplaceAll(filePath, "'", "''")
		cmd = exec.Command("powershell", "-Command",
			"Add-Type -AssemblyName Microsoft.VisualBasic; [Microsoft.VisualBasic.FileIO.FileSystem]::DeleteFile('"+escaped+"', 'OnlyErrorDialogs', 'SendToRecycleBin')")
	case "darwin":
		cmd = exec.Command("mv", filePath, os.Getenv("HOME")+"/.Trash/")
	case "linux":
		cmd = exec.Command("gio", "trash", filePath)
	default:
		return exec.Command("rm", filePath).Run()
	}
	return cmd.Run()
}
