package api

import (
	"catch/internal/application/dto"
	"catch/internal/application/service"
	"catch/internal/domain/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileAppSvc *service.FileAppService
	hub        *ProgressHub
}

func NewFileHandler(fileAppSvc *service.FileAppService) *FileHandler {
	return &FileHandler{
		fileAppSvc: fileAppSvc,
		hub:        GetProgressHub(),
	}
}

func (h *FileHandler) RegisterRoutes(rg *gin.RouterGroup) {
	files := rg.Group("/files")
	{
		files.GET("/search", h.Search)
		files.GET("/browse", h.Browse)
		files.POST("/delete", h.Delete)
		files.POST("/rename", h.Rename)
		files.POST("/rename/preview", h.RenamePreview)
		files.POST("/move", h.Move)
		files.POST("/copy", h.Copy)
	}
}

func (h *FileHandler) Search(c *gin.Context) {
	var req dto.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		path := c.Query("path")
		pattern := c.Query("pattern")
		fileType := c.Query("file_type")
		req = dto.SearchRequest{
			Path:     path,
			Pattern:  pattern,
			FileType: fileType,
		}
	}

	clientID := c.Query("client_id")

	var progressCb entity.ProgressCallback
	if clientID != "" {
		progressCb = func(progress entity.SearchProgress) {
			h.hub.BroadcastSearchProgress(clientID, progress.Scanned, progress.Found, progress.CurrentDir)
		}
	}

	resp, err := h.fileAppSvc.Search(req, progressCb)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) Browse(c *gin.Context) {
	path := c.Query("path")
	resp, err := h.fileAppSvc.Browse(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) Delete(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientID := c.GetHeader("X-Client-ID")

	var progressCb func(done int, total int)
	if clientID != "" {
		progressCb = func(done int, total int) {
			h.hub.BroadcastOperationProgress(clientID, "delete", done, total)
		}
	}

	resp, err := h.fileAppSvc.Delete(req, progressCb)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) RenamePreview(c *gin.Context) {
	var req dto.RenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.fileAppSvc.RenamePreview(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) Rename(c *gin.Context) {
	var req dto.RenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.fileAppSvc.Rename(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) Move(c *gin.Context) {
	var req dto.MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.fileAppSvc.Move(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) Copy(c *gin.Context) {
	var req dto.CopyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.fileAppSvc.Copy(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
