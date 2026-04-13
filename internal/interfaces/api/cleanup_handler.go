package api

import (
	"catch/internal/application/dto"
	"catch/internal/application/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CleanupHandler struct {
	cleanupAppSvc *service.CleanupAppService
	hub           *ProgressHub
}

func NewCleanupHandler(cleanupAppSvc *service.CleanupAppService) *CleanupHandler {
	return &CleanupHandler{
		cleanupAppSvc: cleanupAppSvc,
		hub:           GetProgressHub(),
	}
}

func (h *CleanupHandler) RegisterRoutes(rg *gin.RouterGroup) {
	cleanup := rg.Group("/cleanup")
	{
		cleanup.GET("/rules", h.GetRules)
		cleanup.POST("/scan", h.Scan)
		cleanup.POST("/execute", h.Execute)
	}
}

func (h *CleanupHandler) GetRules(c *gin.Context) {
	resp := h.cleanupAppSvc.GetRules()
	c.JSON(http.StatusOK, resp)
}

func (h *CleanupHandler) Scan(c *gin.Context) {
	var req dto.CleanupScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientID := c.Query("client_id")

	var progressCb func(done int, total int)
	if clientID != "" {
		progressCb = func(done int, total int) {
			h.hub.BroadcastOperationProgress(clientID, "cleanup_scan", done, total)
		}
	}

	resp, err := h.cleanupAppSvc.Scan(req, progressCb)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *CleanupHandler) Execute(c *gin.Context) {
	var req dto.CleanupExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientID := c.GetHeader("X-Client-ID")

	var progressCb func(done int, total int)
	if clientID != "" {
		progressCb = func(done int, total int) {
			h.hub.BroadcastOperationProgress(clientID, "cleanup_execute", done, total)
		}
	}

	resp, err := h.cleanupAppSvc.Execute(req, progressCb)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
