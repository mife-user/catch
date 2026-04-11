package persistence

import (
	"catch/internal/domain/entity"
	"catch/internal/domain/repository"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileRepositoryImpl struct {
	mu sync.RWMutex
}

func NewFileRepository() repository.FileRepository {
	return &FileRepositoryImpl{}
}

func (r *FileRepositoryImpl) Search(path string, pattern string, fileType entity.FileType, customExts []string, minSize int64, maxSize int64, modAfter string, modBefore string) ([]*entity.FileInfo, []string, error) {
	var files []*entity.FileInfo
	var skipped []string

	extensions := entity.GetExtensionsForType(fileType, customExts)

	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			skipped = append(skipped, currentPath+": "+err.Error())
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if pattern != "" {
			matched, err := filepath.Match(pattern, info.Name())
			if err != nil || !matched {
				matched, err = filepath.Match("*"+pattern+"*", info.Name())
				if err != nil || !matched {
					return nil
				}
			}
		}

		if extensions != nil {
			ext := filepath.Ext(info.Name())
			found := false
			for _, allowedExt := range extensions {
				if ext == allowedExt {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		if minSize > 0 && info.Size() < minSize {
			return nil
		}
		if maxSize > 0 && info.Size() > maxSize {
			return nil
		}

		if modAfter != "" || modBefore != "" {
			modTime := info.ModTime()
			if modAfter != "" {
				after, err := time.Parse("2006-01-02", modAfter)
				if err == nil && modTime.Before(after) {
					return nil
				}
			}
			if modBefore != "" {
				before, err := time.Parse("2006-01-02", modBefore)
				if err == nil && modTime.After(before) {
					return nil
				}
			}
		}

		files = append(files, entity.NewFileInfoFromOS(currentPath, info))
		return nil
	})

	if err != nil {
		return nil, skipped, fmt.Errorf("遍历目录失败: %w", err)
	}

	return files, skipped, nil
}

func (r *FileRepositoryImpl) Delete(path string) error {
	return os.Remove(path)
}

func (r *FileRepositoryImpl) Move(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	if linkErr, ok := err.(*os.LinkError); ok {
		if r.Copy(src, dst) != nil {
			return linkErr
		}
		if removeErr := os.Remove(src); removeErr != nil {
			return removeErr
		}
		return nil
	}

	return err
}

func (r *FileRepositoryImpl) Copy(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		os.Remove(dst)
		return err
	}

	srcInfo, err := os.Stat(src)
	if err == nil {
		os.Chmod(dst, srcInfo.Mode())
	}

	return nil
}

func (r *FileRepositoryImpl) Rename(oldPath, newPath string) error {
	if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
		return err
	}
	return os.Rename(oldPath, newPath)
}

func (r *FileRepositoryImpl) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
