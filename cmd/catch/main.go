package main

import (
	"catch/internal/application/service"
	domainService "catch/internal/domain/service"
	"catch/internal/infrastructure/browser"
	"catch/internal/infrastructure/persistence"
	"catch/internal/interfaces/api"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Catch 文件整理工具 v1.0.0")
	fmt.Println("===============================")
	fmt.Println("正在启动服务...")

	configRepo := persistence.NewConfigRepository()
	fileRepo := persistence.NewFileRepository()
	trashRepo := persistence.NewTrashRepository()

	configAppSvc := service.NewConfigAppService(configRepo)

	if err := configAppSvc.EnsureConfig(); err != nil {
		fmt.Printf("初始化配置失败: %v\n", err)
		os.Exit(1)
	}

	fileDomainSvc := domainService.NewFileDomainService(fileRepo)
	trashDomainSvc := domainService.NewTrashDomainService(trashRepo, configRepo)

	fileAppSvc := service.NewFileAppService(fileRepo, configRepo, trashRepo, fileDomainSvc, trashDomainSvc)
	feedbackAppSvc := service.NewFeedbackAppService(configRepo)
	trashAppSvc := service.NewTrashAppService(trashRepo, configRepo, trashDomainSvc)

	if err := trashDomainSvc.StartupCleanup(); err != nil {
		fmt.Printf("启动清理过期文件失败: %v\n", err)
	}

	port, err := browser.FindAvailablePort(3000, 3100)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("检测到可用端口: %d\n", port)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	router := api.NewRouter(fileAppSvc, configAppSvc, feedbackAppSvc, trashAppSvc)
	router.Setup(engine)

	staticPath := findStaticFiles()
	if staticPath != "" {
		engine.Static("/assets", filepath.Join(staticPath, "assets"))
		engine.StaticFile("/favicon.svg", filepath.Join(staticPath, "favicon.svg"))
		engine.NoRoute(func(c *gin.Context) {
			indexPath := filepath.Join(staticPath, "index.html")
			c.File(indexPath)
		})
		fmt.Println("已加载前端静态资源")
	} else {
		fmt.Println("未找到前端静态资源，仅API模式运行")
	}

	addr := fmt.Sprintf(":%d", port)
	url := fmt.Sprintf("http://localhost:%d", port)
	fmt.Printf("服务已启动: %s\n", url)
	fmt.Println("正在打开浏览器...")

	if err := browser.Open(url); err != nil {
		fmt.Printf("无法自动打开浏览器，请手动访问: %s\n", url)
	}

	fmt.Println("按 Ctrl+C 停止服务")

	server := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("服务启动失败: %v\n", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n正在停止服务...")
	fmt.Println("服务已停止")
}

func findStaticFiles() string {
	exePath, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exePath), "web", "dist")
		if isDir(candidate) {
			return candidate
		}
	}

	candidate := filepath.Join(".", "web", "dist")
	if isDir(candidate) {
		return candidate
	}

	wd, err := os.Getwd()
	if err == nil {
		candidate = filepath.Join(wd, "web", "dist")
		if isDir(candidate) {
			return candidate
		}
	}

	return ""
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
