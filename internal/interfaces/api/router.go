package api

import (
	"catch/internal/application/service"
	"github.com/gin-gonic/gin"
)

type Router struct {
	fileHandler     *FileHandler
	configHandler   *ConfigHandler
	feedbackHandler *FeedbackHandler
	trashHandler    *TrashHandler
}

func NewRouter(
	fileAppSvc *service.FileAppService,
	configAppSvc *service.ConfigAppService,
	feedbackAppSvc *service.FeedbackAppService,
	trashAppSvc *service.TrashAppService,
) *Router {
	return &Router{
		fileHandler:     NewFileHandler(fileAppSvc),
		configHandler:   NewConfigHandler(configAppSvc),
		feedbackHandler: NewFeedbackHandler(feedbackAppSvc),
		trashHandler:    NewTrashHandler(trashAppSvc),
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	api := engine.Group("/api")
	{
		r.fileHandler.RegisterRoutes(api)
		r.configHandler.RegisterRoutes(api)
		r.feedbackHandler.RegisterRoutes(api)
		r.trashHandler.RegisterRoutes(api)
	}
}
