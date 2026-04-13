package api

import (
	"catch/internal/application/service"

	"github.com/gin-gonic/gin"
)

type Router struct {
	fileHandler      *FileHandler
	configHandler    *ConfigHandler
	feedbackHandler  *FeedbackHandler
	trashHandler     *TrashHandler
	wsHandler        *WebSocketHandler
	cleanupHandler   *CleanupHandler
	uninstallHandler *UninstallHandler
}

func NewRouter(
	fileAppSvc *service.FileAppService,
	configAppSvc *service.ConfigAppService,
	feedbackAppSvc *service.FeedbackAppService,
	trashAppSvc *service.TrashAppService,
	cleanupAppSvc *service.CleanupAppService,
	uninstallAppSvc *service.UninstallAppService,
) *Router {
	return &Router{
		fileHandler:      NewFileHandler(fileAppSvc),
		configHandler:    NewConfigHandler(configAppSvc),
		feedbackHandler:  NewFeedbackHandler(feedbackAppSvc),
		trashHandler:     NewTrashHandler(trashAppSvc),
		wsHandler:        NewWebSocketHandler(),
		cleanupHandler:   NewCleanupHandler(cleanupAppSvc),
		uninstallHandler: NewUninstallHandler(uninstallAppSvc),
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	api := engine.Group("/api")
	{
		r.fileHandler.RegisterRoutes(api)
		r.configHandler.RegisterRoutes(api)
		r.feedbackHandler.RegisterRoutes(api)
		r.trashHandler.RegisterRoutes(api)
		r.wsHandler.RegisterRoutes(api)
		r.cleanupHandler.RegisterRoutes(api)
		r.uninstallHandler.RegisterRoutes(api)
	}
}
