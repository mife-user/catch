package api

import (
	"catch/internal/application/dto"
	"catch/internal/application/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileAppSvc *service.FileAppService
}

func NewFileHandler(fileAppSvc *service.FileAppService) *FileHandler {
	return &FileHandler{fileAppSvc: fileAppSvc}
}

func (h *FileHandler) RegisterRoutes(rg *gin.RouterGroup) {
	files := rg.Group("/files")
	{
		files.GET("/search", h.Search)
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

	resp, err := h.fileAppSvc.Search(req)
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

	resp, err := h.fileAppSvc.Delete(req)
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
