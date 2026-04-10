package dto

type TrashItemResponse struct {
	OriginalPath  string `json:"original_path"`
	TrashPath     string `json:"trash_path"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	DeletedAt     string `json:"deleted_at"`
	ExpiresAt     string `json:"expires_at"`
	RemainingDays int    `json:"remaining_days"`
	IsExpired     bool   `json:"is_expired"`
}

type TrashListResponse struct {
	Items []TrashItemResponse `json:"items"`
	Total int                 `json:"total"`
}

type TrashRestoreRequest struct {
	OriginalPaths []string `json:"original_paths"`
}

type TrashRestoreResponse struct {
	Success []string `json:"success"`
	Failed  []string `json:"failed"`
}

type TrashCleanResponse struct {
	Cleaned int `json:"cleaned"`
}
