package persistence

import (
	"catch/internal/domain/entity"
	"catch/internal/domain/repository"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type FileRepositoryImpl struct {
	mu sync.RWMutex
}

func NewFileRepository() repository.FileRepository {
	return &FileRepositoryImpl{}
}

func (r *FileRepositoryImpl) Search(rootPath string, pattern string, fileType entity.FileType, customExts []string, minSize int64, maxSize int64, modAfter string, modBefore string, progressCb entity.ProgressCallback) ([]*entity.FileInfo, []string, error) {
	extensions := entity.GetExtensionsForType(fileType, customExts)

	var filesMu sync.Mutex
	var files []*entity.FileInfo
	var skipped []string
	var scannedCount int64
	var foundCount int64

	dirs := r.collectSubDirs(rootPath)

	numWorkers := 4
	if len(dirs) < numWorkers {
		numWorkers = len(dirs)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	dirCh := make(chan string, len(dirs)+1)
	dirCh <- rootPath
	for _, d := range dirs {
		dirCh <- d
	}
	close(dirCh)

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dir := range dirCh {
				r.walkDir(dir, pattern, extensions, minSize, maxSize, modAfter, modBefore, &filesMu, &files, &skipped, &scannedCount, &foundCount, progressCb)
			}
		}()
	}

	wg.Wait()

	if progressCb != nil {
		progressCb(entity.SearchProgress{
			Scanned:    int(atomic.LoadInt64(&scannedCount)),
			Found:      int(atomic.LoadInt64(&foundCount)),
			CurrentDir: rootPath,
		})
	}

	return files, skipped, nil
}

func (r *FileRepositoryImpl) collectSubDirs(rootPath string) []string {
	var dirs []string

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return dirs
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if len(entry.Name()) > 0 && entry.Name()[0] == '.' {
			continue
		}
		fullPath := filepath.Join(rootPath, entry.Name())
		dirs = append(dirs, fullPath)
	}

	return dirs
}

func (r *FileRepositoryImpl) walkDir(dir string, pattern string, extensions []string, minSize int64, maxSize int64, modAfter string, modBefore string, filesMu *sync.Mutex, files *[]*entity.FileInfo, skipped *[]string, scannedCount *int64, foundCount *int64, progressCb entity.ProgressCallback) {
	filepath.Walk(dir, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			filesMu.Lock()
			*skipped = append(*skipped, currentPath+": "+err.Error())
			filesMu.Unlock()
			return nil
		}

		if info.IsDir() {
			return nil
		}

		atomic.AddInt64(scannedCount, 1)

		if pattern != "" {
			matched, matchErr := filepath.Match(pattern, info.Name())
			if matchErr != nil || !matched {
				matched, matchErr = filepath.Match("*"+pattern+"*", info.Name())
				if matchErr != nil || !matched {
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
				after, parseErr := time.Parse("2006-01-02", modAfter)
				if parseErr == nil && modTime.Before(after) {
					return nil
				}
			}
			if modBefore != "" {
				before, parseErr := time.Parse("2006-01-02", modBefore)
				if parseErr == nil && modTime.After(before) {
					return nil
				}
			}
		}

		fileInfo := entity.NewFileInfoFromOS(currentPath, info)
		filesMu.Lock()
		*files = append(*files, fileInfo)
		filesMu.Unlock()

		newFound := atomic.AddInt64(foundCount, 1)

		if progressCb != nil && newFound%50 == 0 {
			progressCb(entity.SearchProgress{
				Scanned:    int(atomic.LoadInt64(scannedCount)),
				Found:      int(newFound),
				CurrentDir: dir,
			})
		}

		return nil
	})
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
