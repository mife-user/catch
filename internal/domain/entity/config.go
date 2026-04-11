package entity

type ServerConfig struct {
	Port int `json:"port"`
}

type TrashConfig struct {
	ExpireDays int    `json:"expire_days"`
	Path       string `json:"path"`
}

type SecurityConfig struct {
	Password     string `json:"password"`
	PasswordHint string `json:"password_hint"`
}

type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	To       string `json:"to"`
}

type SearchConfig struct {
	DefaultPath string `json:"default_path"`
}

type AppConfig struct {
	Version     string         `json:"version"`
	FirstLaunch bool           `json:"first_launch"`
	Server      ServerConfig   `json:"server"`
	Trash       TrashConfig    `json:"trash"`
	Security    SecurityConfig `json:"security"`
	SMTP        SMTPConfig     `json:"smtp"`
	Favorites   []string       `json:"favorites"`
	Search      SearchConfig   `json:"search"`
}

func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		Version:     "1.0.0",
		FirstLaunch: true,
		Server: ServerConfig{
			Port: 3000,
		},
		Trash: TrashConfig{
			ExpireDays: 7,
			Path:       "~/.catch-trash",
		},
		Security: SecurityConfig{
			Password:     "",
			PasswordHint: "",
		},
		SMTP: SMTPConfig{
			Host:     "",
			Port:     465,
			Username: "",
			Password: "",
			To:       "15723556393@163.com",
		},
		Favorites: []string{},
		Search: SearchConfig{
			DefaultPath: "~",
		},
	}
}

func (c *AppConfig) HasPassword() bool {
	return c.Security.Password != ""
}

func (c *AppConfig) HasSMTP() bool {
	return c.SMTP.Host != "" && c.SMTP.Username != "" && c.SMTP.Password != ""
}

func (c *AppConfig) ValidatePassword(password string) bool {
	return c.Security.Password == password
}
