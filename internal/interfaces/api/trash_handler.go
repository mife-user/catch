package api

import (
	"catch/internal/application/dto"
	"catch/internal/application/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TrashHandler struct {
	trashAppSvc *service.TrashAppService
}

func NewTrashHandler(trashAppSvc *service.TrashAppService) *TrashHandler {
	return &TrashHandler{trashAppSvc: trashAppSvc}
}

func (h *TrashHandler) RegisterRoutes(rg *gin.RouterGroup) {
	trash := rg.Group("/trash")
	{
		trash.GET("", h.List)
		trash.POST("/restore", h.Restore)
		trash.DELETE("/clean", h.CleanExpired)
	}
}

func (h *TrashHandler) List(c *gin.Context) {
	resp, err := h.trashAppSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TrashHandler) Restore(c *gin.Context) {
	var req dto.TrashRestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.trashAppSvc.Restore(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TrashHandler) CleanExpired(c *gin.Context) {
	resp, err := h.trashAppSvc.CleanExpired()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
