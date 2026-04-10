package entity

import (
	"os"
	"path/filepath"
	"time"
)

type FileInfo struct {
	Name         string      `json:"name"`
	Path         string      `json:"path"`
	Size         int64       `json:"size"`
	ModTime      time.Time   `json:"mod_time"`
	IsDir        bool        `json:"is_dir"`
	Extension    string      `json:"extension"`
	Permissions  os.FileMode `json:"permissions"`
}

func NewFileInfoFromOS(path string, info os.FileInfo) *FileInfo {
	return &FileInfo{
		Name:        info.Name(),
		Path:        path,
		Size:        info.Size(),
		ModTime:     info.ModTime(),
		IsDir:       info.IsDir(),
		Extension:   filepath.Ext(info.Name()),
		Permissions: info.Mode(),
	}
}

func (f *FileInfo) IsReadable() bool {
	file, err := os.Open(f.Path)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

type FileType string

const (
	FileTypeAll      FileType = "all"
	FileTypeDocument FileType = "document"
	FileTypeImage    FileType = "image"
	FileTypeVideo    FileType = "video"
	FileTypeAudio    FileType = "audio"
	FileTypeCustom   FileType = "custom"
)

var FileTypeExtensions = map[FileType][]string{
	FileTypeDocument: {".txt", ".doc", ".docx", ".pdf", ".xls", ".xlsx", ".ppt", ".pptx", ".csv", ".md", ".rtf"},
	FileTypeImage:    {".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp", ".ico", ".tiff"},
	FileTypeVideo:    {".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm"},
	FileTypeAudio:    {".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a"},
}

func GetExtensionsForType(fileType FileType, customExts []string) []string {
	if fileType == FileTypeAll {
		return nil
	}
	if fileType == FileTypeCustom {
		return customExts
	}
	return FileTypeExtensions[fileType]
}
