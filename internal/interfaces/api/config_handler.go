package api

import (
	"catch/internal/application/dto"
	"catch/internal/application/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configAppSvc *service.ConfigAppService
}

func NewConfigHandler(configAppSvc *service.ConfigAppService) *ConfigHandler {
	return &ConfigHandler{configAppSvc: configAppSvc}
}

func (h *ConfigHandler) RegisterRoutes(rg *gin.RouterGroup) {
	config := rg.Group("/config")
	{
		config.GET("", h.GetConfig)
		config.PUT("", h.UpdateConfig)
		config.POST("/password", h.SetPassword)
		config.POST("/password/verify", h.VerifyPassword)
		config.DELETE("/password", h.RemovePassword)
	}
}

func (h *ConfigHandler) GetConfig(c *gin.Context) {
	resp, err := h.configAppSvc.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	var req dto.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.configAppSvc.UpdateConfig(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) SetPassword(c *gin.Context) {
	var req dto.SetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.configAppSvc.SetPassword(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码设置成功"})
}

func (h *ConfigHandler) VerifyPassword(c *gin.Context) {
	var req dto.VerifyPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid := h.configAppSvc.VerifyPassword(req)
	c.JSON(http.StatusOK, gin.H{"valid": valid})
}

func (h *ConfigHandler) RemovePassword(c *gin.Context) {
	var req dto.RemovePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.configAppSvc.RemovePassword(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码已删除"})
}
