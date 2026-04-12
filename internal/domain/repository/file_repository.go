package repository

import "catch/internal/domain/entity"

type FileRepository interface {
	Search(path string, pattern string, fileType entity.FileType, customExts []string, minSize int64, maxSize int64, modAfter string, modBefore string, progressCb entity.ProgressCallback) ([]*entity.FileInfo, []string, error)
	Delete(path string) error
	Move(src, dst string) error
	Copy(src, dst string) error
	Rename(oldPath, newPath string) error
	Exists(path string) bool
}
