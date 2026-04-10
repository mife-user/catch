package service

import (
	"os"
	"os/exec"
	"runtime"
)

func moveToSystemTrash(filePath string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("powershell", "-Command", "Add-Type -AssemblyName Microsoft.VisualBasic; [Microsoft.VisualBasic.FileIO.FileSystem]::DeleteFile('"+filePath+"', 'OnlyErrorDialogs', 'SendToRecycleBin')")
	case "darwin":
		cmd = exec.Command("mv", filePath, os.Getenv("HOME")+"/.Trash/")
	case "linux":
		cmd = exec.Command("gio", "trash", filePath)
	default:
		return exec.Command("rm", filePath).Run()
	}
	return cmd.Run()
}
