package api

import (
	"catch/internal/application/dto"
	"catch/internal/application/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UninstallHandler struct {
	uninstallAppSvc *service.UninstallAppService
	hub             *ProgressHub
}

func NewUninstallHandler(uninstallAppSvc *service.UninstallAppService) *UninstallHandler {
	return &UninstallHandler{
		uninstallAppSvc: uninstallAppSvc,
		hub:             GetProgressHub(),
	}
}

func (h *UninstallHandler) RegisterRoutes(rg *gin.RouterGroup) {
	uninstall := rg.Group("/uninstall")
	{
		uninstall.GET("/scan", h.Scan)
		uninstall.POST("/analyze", h.Analyze)
		uninstall.POST("/execute", h.Execute)
	}
}

func (h *UninstallHandler) Scan(c *gin.Context) {
	resp, err := h.uninstallAppSvc.Scan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UninstallHandler) Analyze(c *gin.Context) {
	var req dto.UninstallAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.uninstallAppSvc.Analyze(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UninstallHandler) Execute(c *gin.Context) {
	var req dto.UninstallExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientID := c.GetHeader("X-Client-ID")

	var progressCb func(done int, total int)
	if clientID != "" {
		progressCb = func(done int, total int) {
			h.hub.BroadcastOperationProgress(clientID, "uninstall", done, total)
		}
	}

	resp, err := h.uninstallAppSvc.Execute(req, progressCb)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
