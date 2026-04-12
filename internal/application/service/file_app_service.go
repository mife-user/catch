package service

import (
	"catch/internal/application/dto"
	"catch/internal/domain/entity"
	"catch/internal/domain/repository"
	"catch/internal/domain/service"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileAppService struct {
	fileRepo       repository.FileRepository
	configRepo     repository.ConfigRepository
	trashRepo      repository.TrashRepository
	fileDomainSvc  *service.FileDomainService
	trashDomainSvc *service.TrashDomainService
}

func NewFileAppService(
	fileRepo repository.FileRepository,
	configRepo repository.ConfigRepository,
	trashRepo repository.TrashRepository,
	fileDomainSvc *service.FileDomainService,
	trashDomainSvc *service.TrashDomainService,
) *FileAppService {
	return &FileAppService{
		fileRepo:       fileRepo,
		configRepo:     configRepo,
		trashRepo:      trashRepo,
		fileDomainSvc:  fileDomainSvc,
		trashDomainSvc: trashDomainSvc,
	}
}

func (s *FileAppService) Search(req dto.SearchRequest, progressCb entity.ProgressCallback) (*dto.SearchResponse, error) {
	fileType := entity.FileType(req.FileType)
	if fileType == "" {
		fileType = entity.FileTypeAll
	}

	if req.Path == "" {
		config, err := s.configRepo.Load()
		if err != nil {
			return nil, fmt.Errorf("无法加载配置: %w", err)
		}
		req.Path = config.Search.DefaultPath
		if req.Path == "~" {
			home, _ := os.UserHomeDir()
			req.Path = home
		}
	}

	files, skipped, err := s.fileDomainSvc.SearchFiles(req.Path, req.Pattern, fileType, req.CustomExts, req.MinSize, req.MaxSize, req.ModAfter, req.ModBefore, progressCb)
	if err != nil {
		return nil, fmt.Errorf("文件查找失败: %w", err)
	}

	results := make([]dto.FileResult, 0, len(files))
	for _, f := range files {
		results = append(results, dto.FileResult{
			Name:        f.Name,
			Path:        f.Path,
			Size:        f.Size,
			ModTime:     f.ModTime.Format("2006-01-02 15:04:05"),
			IsDir:       f.IsDir,
			Extension:   f.Extension,
			Permissions: f.Permissions.String(),
		})
	}

	return &dto.SearchResponse{
		Files:        results,
		Skipped:      skipped,
		Total:        len(results),
		SkippedCount: len(skipped),
	}, nil
}

func (s *FileAppService) Delete(req dto.DeleteRequest, progressCb func(done int, total int)) (*dto.DeleteResponse, error) {
	config, err := s.configRepo.Load()
	if err != nil {
		return nil, fmt.Errorf("无法加载配置: %w", err)
	}

	success := make([]string, 0)
	failed := make([]string, 0)
	total := len(req.Paths)

	switch req.Mode {
	case "recycle":
		for i, path := range req.Paths {
			if err := s.fileDomainSvc.MoveToSystemTrash(path); err != nil {
				failed = append(failed, path+": "+err.Error())
			} else {
				success = append(success, path)
			}
			if progressCb != nil {
				progressCb(i+1, total)
			}
		}

	case "trash":
		trashDir, err := s.trashDomainSvc.GetTrashDir()
		if err != nil {
			return nil, fmt.Errorf("无法获取回收站目录: %w", err)
		}

		for i, path := range req.Paths {
			trashPath := s.fileDomainSvc.GenerateTrashPath(trashDir, path)
			item, err := s.fileDomainSvc.MoveToTrash(path, trashPath, config.Trash.ExpireDays)
			if err != nil {
				failed = append(failed, path+": "+err.Error())
				if progressCb != nil {
					progressCb(i+1, total)
				}
				continue
			}
			if err := s.trashRepo.Add(item); err != nil {
				failed = append(failed, path+": 保存回收站记录失败")
				if progressCb != nil {
					progressCb(i+1, total)
				}
				continue
			}
			success = append(success, path)
			if progressCb != nil {
				progressCb(i+1, total)
			}
		}

	case "permanent":
		if !config.HasPassword() {
			return nil, fmt.Errorf("未设置安全密码，无法使用永久删除功能")
		}
		if !config.ValidatePassword(req.Password) {
			return nil, fmt.Errorf("安全密码验证失败")
		}

		for i, path := range req.Paths {
			if err := s.fileDomainSvc.PermanentDelete(path); err != nil {
				failed = append(failed, path+": "+err.Error())
			} else {
				success = append(success, path)
			}
			if progressCb != nil {
				progressCb(i+1, total)
			}
		}

	default:
		return nil, fmt.Errorf("不支持的删除模式: %s", req.Mode)
	}

	return &dto.DeleteResponse{Success: success, Failed: failed}, nil
}

func (s *FileAppService) RenamePreview(req dto.RenameRequest) (*dto.RenameResponse, error) {
	files := make([]*entity.FileInfo, 0, len(req.Paths))
	for _, p := range req.Paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		files = append(files, entity.NewFileInfoFromOS(p, info))
	}

	renameMap, err := s.fileDomainSvc.GenerateRenamePattern(files, req.Rule, req.Params)
	if err != nil {
		return nil, err
	}

	previews := make([]dto.RenamePreview, 0, len(renameMap))
	for oldPath, newPath := range renameMap {
		previews = append(previews, dto.RenamePreview{OldPath: oldPath, NewPath: newPath})
	}

	return &dto.RenameResponse{Previews: previews}, nil
}

func (s *FileAppService) Rename(req dto.RenameRequest) (*dto.RenameResponse, error) {
	resp, err := s.RenamePreview(req)
	if err != nil {
		return nil, err
	}

	success := make([]string, 0)
	failed := make([]string, 0)

	for _, preview := range resp.Previews {
		if err := s.fileRepo.Rename(preview.OldPath, preview.NewPath); err != nil {
			failed = append(failed, preview.OldPath+": "+err.Error())
		} else {
			success = append(success, preview.OldPath)
		}
	}

	resp.Success = success
	resp.Failed = failed
	return resp, nil
}

func (s *FileAppService) Move(req dto.MoveRequest) (*dto.MoveResponse, error) {
	return s.moveOrCopy(req.SrcPaths, req.DstPath, req.Conflict, "move")
}

func (s *FileAppService) Copy(req dto.CopyRequest) (*dto.CopyResponse, error) {
	resp, err := s.moveOrCopy(req.SrcPaths, req.DstPath, req.Conflict, "copy")
	if err != nil {
		return nil, err
	}
	return &dto.CopyResponse{Success: resp.Success, Failed: resp.Failed, Skipped: resp.Skipped}, nil
}

func (s *FileAppService) moveOrCopy(srcPaths []string, dstPath string, conflict string, op string) (*dto.MoveResponse, error) {
	if err := os.MkdirAll(dstPath, 0755); err != nil {
		return nil, fmt.Errorf("无法创建目标目录: %w", err)
	}

	success := make([]string, 0)
	failed := make([]string, 0)
	skipped := make([]string, 0)

	var mu sync.Mutex
	var wg sync.WaitGroup

	numWorkers := 4
	if len(srcPaths) < numWorkers {
		numWorkers = len(srcPaths)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	pathCh := make(chan int, len(srcPaths))
	for i := range srcPaths {
		pathCh <- i
	}
	close(pathCh)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range pathCh {
				src := srcPaths[idx]
				fileName := filepath.Base(src)
				dst := filepath.Join(dstPath, fileName)

				if _, err := os.Stat(dst); err == nil {
					switch conflict {
					case "skip":
						mu.Lock()
						skipped = append(skipped, src)
						mu.Unlock()
						continue
					case "rename":
						ext := filepath.Ext(fileName)
						base := strings.TrimSuffix(fileName, ext)
						counter := 1
						for {
							dst = filepath.Join(dstPath, fmt.Sprintf("%s_(%d)%s", base, counter, ext))
							if _, err := os.Stat(dst); os.IsNotExist(err) {
								break
							}
							counter++
						}
					case "overwrite":
						os.Remove(dst)
					default:
						mu.Lock()
						skipped = append(skipped, src)
						mu.Unlock()
						continue
					}
				}

				var err error
				if op == "move" {
					err = s.fileRepo.Move(src, dst)
				} else {
					err = s.fileRepo.Copy(src, dst)
				}

				mu.Lock()
				if err != nil {
					failed = append(failed, src+": "+err.Error())
				} else {
					success = append(success, src)
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	return &dto.MoveResponse{Success: success, Failed: failed, Skipped: skipped}, nil
}

func (s *FileAppService) Browse(path string) (*dto.BrowseResponse, error) {
	if path == "" || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("无法获取用户主目录: %w", err)
		}
		path = home
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("无效路径: %w", err)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("无法读取目录: %w", err)
	}

	parentPath := filepath.Dir(absPath)
	items := make([]dto.BrowseItem, 0)

	for _, entry := range entries {
		fullPath := filepath.Join(absPath, entry.Name())
		_, err := entry.Info()
		if err != nil {
			continue
		}

		if !entry.IsDir() {
			continue
		}

		if len(entry.Name()) > 0 && entry.Name()[0] == '.' {
			continue
		}

		items = append(items, dto.BrowseItem{
			Name:  entry.Name(),
			Path:  fullPath,
			IsDir: true,
		})
	}

	return &dto.BrowseResponse{
		CurrentPath: absPath,
		ParentPath:  parentPath,
		Items:       items,
	}, nil
}
