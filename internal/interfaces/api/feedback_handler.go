package api

import (
	"catch/internal/application/dto"
	"catch/internal/application/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FeedbackHandler struct {
	feedbackAppSvc *service.FeedbackAppService
}

func NewFeedbackHandler(feedbackAppSvc *service.FeedbackAppService) *FeedbackHandler {
	return &FeedbackHandler{feedbackAppSvc: feedbackAppSvc}
}

func (h *FeedbackHandler) RegisterRoutes(rg *gin.RouterGroup) {
	feedback := rg.Group("/feedback")
	{
		feedback.POST("", h.SendFeedback)
	}

	smtp := rg.Group("/smtp")
	{
		smtp.GET("/templates", h.GetSMTPTemplates)
		smtp.POST("/test", h.TestSMTP)
	}
}

func (h *FeedbackHandler) SendFeedback(c *gin.Context) {
	var req dto.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.feedbackAppSvc.SendFeedback(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FeedbackHandler) GetSMTPTemplates(c *gin.Context) {
	templates := h.feedbackAppSvc.GetSMTPTemplates()
	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *FeedbackHandler) TestSMTP(c *gin.Context) {
	var req dto.SMTPTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.feedbackAppSvc.TestSMTP(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
