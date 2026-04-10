package main

import (
	"catch/internal/application/service"
	domainService "catch/internal/domain/service"
	"catch/internal/infrastructure/browser"
	"catch/internal/infrastructure/persistence"
	"catch/internal/interfaces/api"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

const (
	Version   = "1.0.0"
	PortStart = 3000
	PortEnd   = 3100
)

func main() {
	printBanner()

	configRepo := persistence.NewConfigRepository()
	fileRepo := persistence.NewFileRepository()
	trashRepo := persistence.NewTrashRepository()

	configAppSvc := service.NewConfigAppService(configRepo)
	if err := configAppSvc.EnsureConfig(); err != nil {
		log.Printf("警告: 初始化配置文件失败: %v\n", err)
	}

	trashDomainSvc := domainService.NewTrashDomainService(trashRepo, configRepo)
	if err := trashDomainSvc.StartupCleanup(); err != nil {
		log.Printf("警告: 启动时清理过期文件失败: %v\n", err)
	}

	fileDomainSvc := domainService.NewFileDomainService(fileRepo)

	fileAppSvc := service.NewFileAppService(fileRepo, configRepo, trashRepo, fileDomainSvc, trashDomainSvc)
	feedbackAppSvc := service.NewFeedbackAppService(configRepo)
	trashAppSvc := service.NewTrashAppService(trashRepo, configRepo, trashDomainSvc)

	port, err := browser.FindAvailablePort(PortStart, PortEnd)
	if err != nil {
		log.Fatalf("错误: %v\n", err)
	}
	fmt.Printf("检测到可用端口: %d\n", port)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	engine.Use(corsMiddleware())

	router := api.NewRouter(fileAppSvc, configAppSvc, feedbackAppSvc, trashAppSvc)
	router.Setup(engine)

	setupStaticFiles(engine)

	addr := fmt.Sprintf(":%d", port)
	url := fmt.Sprintf("http://localhost:%d", port)

	fmt.Printf("服务已启动: %s\n", url)
	fmt.Println("正在打开浏览器...")

	if err := browser.Open(url); err != nil {
		fmt.Printf("无法自动打开浏览器，请手动访问: %s\n", url)
	}

	fmt.Println("按 Ctrl+C 停止服务")

	go func() {
		if err := engine.Run(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n服务已停止")
}

func printBanner() {
	fmt.Println("Catch 文件整理工具 v" + Version)
	fmt.Println("===============================")
	fmt.Println("正在启动服务...")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func setupStaticFiles(engine *gin.Engine) {
	wd, _ := os.Getwd()
	distPath := wd + "/web/dist"
	if _, err := os.Stat(distPath); err == nil {
		engine.Static("/assets", distPath+"/assets")
		engine.StaticFile("/", distPath+"/index.html")
		engine.StaticFile("/index.html", distPath+"/index.html")
		engine.NoRoute(func(c *gin.Context) {
			c.File(distPath + "/index.html")
		})
	}
}
