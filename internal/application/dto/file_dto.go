package dto

type SearchRequest struct {
	Path       string   `json:"path"`
	Pattern    string   `json:"pattern"`
	FileType   string   `json:"file_type"`
	CustomExts []string `json:"custom_exts"`
	MinSize    int64    `json:"min_size"`
	MaxSize    int64    `json:"max_size"`
	ModAfter   string   `json:"mod_after"`
	ModBefore  string   `json:"mod_before"`
}

type FileResult struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	ModTime     string `json:"mod_time"`
	IsDir       bool   `json:"is_dir"`
	Extension   string `json:"extension"`
	Permissions string `json:"permissions"`
}

type SearchResponse struct {
	Files       []FileResult `json:"files"`
	Skipped     []string     `json:"skipped"`
	Total       int          `json:"total"`
	SkippedCount int         `json:"skipped_count"`
}

type DeleteRequest struct {
	Paths   []string `json:"paths"`
	Mode    string   `json:"mode"`
	Password string  `json:"password,omitempty"`
}

type DeleteResponse struct {
	Success []string `json:"success"`
	Failed  []string `json:"failed"`
}

type RenameRequest struct {
	Paths  []string          `json:"paths"`
	Rule   string            `json:"rule"`
	Params map[string]string `json:"params"`
}

type RenamePreview struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

type RenameResponse struct {
	Previews []RenamePreview `json:"previews"`
	Success  []string        `json:"success"`
	Failed   []string        `json:"failed"`
}

type MoveRequest struct {
	SrcPaths []string `json:"src_paths"`
	DstPath  string   `json:"dst_path"`
	Conflict string   `json:"conflict"`
}

type MoveResponse struct {
	Success []string `json:"success"`
	Failed  []string `json:"failed"`
	Skipped []string `json:"skipped"`
}

type CopyRequest struct {
	SrcPaths []string `json:"src_paths"`
	DstPath  string   `json:"dst_path"`
	Conflict string   `json:"conflict"`
}

type CopyResponse struct {
	Success []string `json:"success"`
	Failed  []string `json:"failed"`
	Skipped []string `json:"skipped"`
}
