package entity

import "time"

type TrashItem struct {
	OriginalPath string    `json:"original_path"`
	TrashPath    string    `json:"trash_path"`
	DeletedAt    time.Time `json:"deleted_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
}

func NewTrashItem(originalPath, trashPath string, fileSize int64, expireDays int) *TrashItem {
	now := time.Now()
	return &TrashItem{
		OriginalPath: originalPath,
		TrashPath:    trashPath,
		DeletedAt:    now,
		ExpiresAt:    now.AddDate(0, 0, expireDays),
		FileName:     extractFileName(originalPath),
		FileSize:     fileSize,
	}
}

func (t *TrashItem) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *TrashItem) CanRestore() bool {
	if t.IsExpired() {
		return false
	}
	return true
}

func extractFileName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '\\' || path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
