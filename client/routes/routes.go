package routes

import (
	"context"
	"gms/client/middleware"
	"gms/client/ui/handlers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Initialize(ctx context.Context, l *zap.Logger) (router *gin.Engine) {
	l.Sugar().Info("Initializing logger")

	router = gin.Default()

	rSecure := router.Group("/sec")
	rSecure.Use(middleware.ContextMiddleware(ctx))
	//TODO:create all other middleware's. set group middleware and chain them together
	rSecure.GET("/home", handlers.HomeHandler)
	rSecure.POST("/checkmail", handlers.CheckMailHandler)
	return router
}
