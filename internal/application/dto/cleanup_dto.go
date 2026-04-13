package dto

type CleanupScanRequest struct {
	Path  string   `json:"path"`
	Rules []string `json:"rules"`
}

type CleanupScanResponse struct {
	Items     []CleanupFileItem `json:"items"`
	Total     int               `json:"total"`
	TotalSize int64             `json:"total_size"`
}

type CleanupFileItem struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	ModTime   string `json:"mod_time"`
	Extension string `json:"extension"`
	Important bool   `json:"important"`
	RuleName  string `json:"rule_name"`
}

type CleanupExecuteRequest struct {
	Paths []string `json:"paths"`
}

type CleanupExecuteResponse struct {
	Cleaned int      `json:"cleaned"`
	Failed  []string `json:"failed"`
	Freed   int64    `json:"freed"`
}

type CleanupRuleItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	FileTypes   []string `json:"file_types"`
	OlderThan   int      `json:"older_than"`
	Important   bool     `json:"important"`
	BuiltIn     bool     `json:"built_in"`
}

type CleanupRulesResponse struct {
	Rules []CleanupRuleItem `json:"rules"`
}

type UninstallApp struct {
	Name         string   `json:"name"`
	InstallPath  string   `json:"install_path"`
	Version      string   `json:"version"`
	Publisher    string   `json:"publisher"`
	RegistryKey  string   `json:"registry_key"`
	IconPath     string   `json:"icon_path"`
	Size         int64    `json:"size"`
	InstallDate  string   `json:"install_date"`
}

type UninstallScanResponse struct {
	Apps []UninstallApp `json:"apps"`
}

type UninstallAnalyzeRequest struct {
	RegistryKey string `json:"registry_key"`
}

type UninstallAnalyzeResponse struct {
	ProgramFiles  []string `json:"program_files"`
	RegistryItems []string `json:"registry_items"`
	ConfigFiles   []string `json:"config_files"`
	ServiceNames  []string `json:"service_names"`
	TotalSize     int64    `json:"total_size"`
	Important     []bool   `json:"important"`
}

type UninstallExecuteRequest struct {
	RegistryKey   string   `json:"registry_key"`
	CleanFiles    []string `json:"clean_files"`
	CleanRegistry []string `json:"clean_registry"`
	CleanConfig   []string `json:"clean_config"`
	CleanServices []string `json:"clean_services"`
}

type UninstallExecuteResponse struct {
	CleanedFiles    int      `json:"cleaned_files"`
	CleanedRegistry int      `json:"cleaned_registry"`
	CleanedServices int      `json:"cleaned_services"`
	Failed          []string `json:"failed"`
	Freed           int64    `json:"freed"`
}
