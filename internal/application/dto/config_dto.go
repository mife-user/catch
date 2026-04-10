package dto

type ConfigResponse struct {
	Version     string        `json:"version"`
	Server      ServerDTO     `json:"server"`
	Trash       TrashDTO      `json:"trash"`
	Security    SecurityDTO   `json:"security"`
	SMTP        SMTPDTO       `json:"smtp"`
	Favorites   []string      `json:"favorites"`
	Search      SearchDTO     `json:"search"`
	HasPassword bool          `json:"has_password"`
	HasSMTP     bool          `json:"has_smtp"`
}

type ServerDTO struct {
	Port int `json:"port"`
}

type TrashDTO struct {
	ExpireDays int    `json:"expire_days"`
	Path       string `json:"path"`
}

type SecurityDTO struct {
	Password     string `json:"password,omitempty"`
	PasswordHint string `json:"password_hint"`
	HasPassword  bool   `json:"has_password"`
}

type SMTPDTO struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	To       string `json:"to"`
	HasSMTP  bool   `json:"has_smtp"`
}

type SearchDTO struct {
	DefaultPath string `json:"default_path"`
}

type UpdateConfigRequest struct {
	Server    *ServerDTO    `json:"server,omitempty"`
	Trash     *TrashDTO     `json:"trash,omitempty"`
	Security  *SecurityDTO  `json:"security,omitempty"`
	SMTP      *SMTPDTO      `json:"smtp,omitempty"`
	Favorites []string      `json:"favorites,omitempty"`
	Search    *SearchDTO    `json:"search,omitempty"`
}

type SetPasswordRequest struct {
	OldPassword string `json:"old_password,omitempty"`
	NewPassword string `json:"new_password"`
	PasswordHint string `json:"password_hint,omitempty"`
}

type VerifyPasswordRequest struct {
	Password string `json:"password"`
}
