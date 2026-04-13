package service

import (
	"catch/internal/application/dto"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type UninstallAppService struct{}

func NewUninstallAppService() *UninstallAppService {
	return &UninstallAppService{}
}

func (s *UninstallAppService) Scan() (*dto.UninstallScanResponse, error) {
	apps := make([]dto.UninstallApp, 0)

	regPaths := []string{
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`HKLM\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, regPath := range regPaths {
		found := s.scanRegistryPath(regPath)
		apps = append(apps, found...)
	}

	seen := make(map[string]bool)
	uniqueApps := make([]dto.UninstallApp, 0)
	for _, app := range apps {
		key := app.Name + "|" + app.InstallPath
		if !seen[key] && app.Name != "" {
			seen[key] = true
			uniqueApps = append(uniqueApps, app)
		}
	}

	return &dto.UninstallScanResponse{Apps: uniqueApps}, nil
}

func (s *UninstallAppService) scanRegistryPath(regPath string) []dto.UninstallApp {
	var apps []dto.UninstallApp

	cmd := exec.Command("reg", "query", regPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return apps
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == regPath {
			continue
		}

		if !strings.HasPrefix(line, "HKEY_") {
			continue
		}

		app := s.readAppFromRegistry(line)
		if app.Name != "" {
			apps = append(apps, app)
		}
	}

	return apps
}

func (s *UninstallAppService) readAppFromRegistry(keyPath string) dto.UninstallApp {
	app := dto.UninstallApp{
		RegistryKey: filepath.Base(keyPath),
	}

	app.Name = s.getRegValue(keyPath, "DisplayName")
	app.InstallPath = s.getRegValue(keyPath, "InstallLocation")
	app.Version = s.getRegValue(keyPath, "DisplayVersion")
	app.Publisher = s.getRegValue(keyPath, "Publisher")
	app.IconPath = s.getRegValue(keyPath, "DisplayIcon")
	app.InstallDate = s.getRegValue(keyPath, "InstallDate")

	if app.InstallPath != "" {
		app.Size = s.calcDirSize(app.InstallPath)
	}

	return app
}

func (s *UninstallAppService) getRegValue(keyPath, valueName string) string {
	cmd := exec.Command("reg", "query", keyPath, "/v", valueName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, valueName) {
			parts := strings.SplitN(line, "REG_SZ", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
			parts = strings.SplitN(line, "REG_EXPAND_SZ", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}

func (s *UninstallAppService) calcDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func (s *UninstallAppService) Analyze(req dto.UninstallAnalyzeRequest) (*dto.UninstallAnalyzeResponse, error) {
	resp := &dto.UninstallAnalyzeResponse{
		ProgramFiles:  []string{},
		RegistryItems: []string{},
		ConfigFiles:   []string{},
		ServiceNames:  []string{},
		Important:     []bool{},
	}

	regPaths := []string{
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`HKLM\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, regPath := range regPaths {
		keyPath := regPath + `\` + req.RegistryKey
		installPath := s.getRegValue(keyPath, "InstallLocation")
		appName := s.getRegValue(keyPath, "DisplayName")

		if installPath != "" {
			s.scanProgramFiles(installPath, resp)
		}

		if appName != "" {
			s.scanRegistryItems(appName, resp)
			s.scanConfigFiles(appName, resp)
			s.scanServices(appName, resp)
		}

		if appName != "" || installPath != "" {
			break
		}
	}

	return resp, nil
}

func (s *UninstallAppService) scanProgramFiles(installPath string, resp *dto.UninstallAnalyzeResponse) {
	filepath.Walk(installPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		resp.ProgramFiles = append(resp.ProgramFiles, path)
		resp.Important = append(resp.Important, false)
		if !info.IsDir() {
			resp.TotalSize += info.Size()
		}
		return nil
	})
}

func (s *UninstallAppService) scanRegistryItems(appName string, resp *dto.UninstallAnalyzeResponse) {
	lowerName := strings.ToLower(appName)

	regPaths := []string{
		`HKLM\SOFTWARE`,
		`HKLM\SOFTWARE\WOW6432Node`,
		`HKCU\SOFTWARE`,
	}

	for _, regPath := range regPaths {
		s.searchRegistryRecursive(regPath, lowerName, resp)
	}
}

func (s *UninstallAppService) searchRegistryRecursive(regPath string, searchName string, resp *dto.UninstallAnalyzeResponse) {
	cmd := exec.Command("reg", "query", regPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(strings.ToLower(line), searchName) {
			resp.RegistryItems = append(resp.RegistryItems, line)
			isImportant := strings.Contains(strings.ToLower(line), "services") ||
				strings.Contains(strings.ToLower(line), "system")
			resp.Important = append(resp.Important, isImportant)
		}
	}
}

func (s *UninstallAppService) scanConfigFiles(appName string, resp *dto.UninstallAnalyzeResponse) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	lowerName := strings.ToLower(strings.ReplaceAll(appName, " ", ""))
	configDirs := []string{
		filepath.Join(home, "AppData", "Roaming"),
		filepath.Join(home, "AppData", "Local"),
		filepath.Join(home, "AppData", "LocalLow"),
	}

	for _, dir := range configDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if strings.Contains(strings.ToLower(entry.Name()), lowerName) {
				fullPath := filepath.Join(dir, entry.Name())
				resp.ConfigFiles = append(resp.ConfigFiles, fullPath)
				resp.Important = append(resp.Important, false)
			}
		}
	}
}

func (s *UninstallAppService) scanServices(appName string, resp *dto.UninstallAnalyzeResponse) {
	cmd := exec.Command("sc", "query", "state=", "all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return
	}

	lowerName := strings.ToLower(appName)
	lines := strings.Split(string(output), "\n")

	var currentServiceName string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SERVICE_NAME:") {
			currentServiceName = strings.TrimPrefix(line, "SERVICE_NAME:")
			currentServiceName = strings.TrimSpace(currentServiceName)
		}
		if strings.HasPrefix(line, "DISPLAY_NAME:") && currentServiceName != "" {
			displayName := strings.TrimPrefix(line, "DISPLAY_NAME:")
			displayName = strings.TrimSpace(displayName)
			if strings.Contains(strings.ToLower(displayName), lowerName) ||
				strings.Contains(strings.ToLower(currentServiceName), lowerName) {
				resp.ServiceNames = append(resp.ServiceNames, currentServiceName)
				resp.Important = append(resp.Important, true)
			}
			currentServiceName = ""
		}
	}
}

func (s *UninstallAppService) Execute(req dto.UninstallExecuteRequest, progressCb func(done int, total int)) (*dto.UninstallExecuteResponse, error) {
	resp := &dto.UninstallExecuteResponse{
		Failed: []string{},
	}

	totalOps := len(req.CleanFiles) + len(req.CleanRegistry) + len(req.CleanConfig) + len(req.CleanServices)
	completed := 0

	for _, path := range req.CleanFiles {
		if err := os.RemoveAll(path); err != nil {
			resp.Failed = append(resp.Failed, "文件: "+path+": "+err.Error())
		} else {
			resp.CleanedFiles++
		}
		completed++
		if progressCb != nil {
			progressCb(completed, totalOps)
		}
	}

	for _, regPath := range req.CleanRegistry {
		if err := s.deleteRegistryKey(regPath); err != nil {
			resp.Failed = append(resp.Failed, "注册表: "+regPath+": "+err.Error())
		} else {
			resp.CleanedRegistry++
		}
		completed++
		if progressCb != nil {
			progressCb(completed, totalOps)
		}
	}

	for _, path := range req.CleanConfig {
		if err := os.RemoveAll(path); err != nil {
			resp.Failed = append(resp.Failed, "配置: "+path+": "+err.Error())
		} else {
			resp.CleanedFiles++
		}
		completed++
		if progressCb != nil {
			progressCb(completed, totalOps)
		}
	}

	for _, svcName := range req.CleanServices {
		if err := s.stopAndDeleteService(svcName); err != nil {
			resp.Failed = append(resp.Failed, "服务: "+svcName+": "+err.Error())
		} else {
			resp.CleanedServices++
		}
		completed++
		if progressCb != nil {
			progressCb(completed, totalOps)
		}
	}

	return resp, nil
}

func (s *UninstallAppService) deleteRegistryKey(path string) error {
	cmd := exec.Command("reg", "delete", path, "/f")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("删除注册表失败: %s: %w", string(output), err)
	}
	return nil
}

func (s *UninstallAppService) stopAndDeleteService(name string) error {
	cmd := exec.Command("net", "stop", name)
	_ = cmd.Run()

	cmd = exec.Command("sc", "delete", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("无法删除服务: %w", err)
	}
	return nil
}
