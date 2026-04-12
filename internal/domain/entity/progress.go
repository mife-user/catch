package entity

type SearchProgress struct {
	Scanned    int    `json:"scanned"`
	Found      int    `json:"found"`
	CurrentDir string `json:"current_dir"`
}

type ProgressCallback func(progress SearchProgress)
