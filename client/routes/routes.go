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
	router.Use(gin.Recovery())

	//secure group
	rSecure := router.Group("/sec")

	rSecure.Use(middleware.ContextMiddleware(ctx))
	rSecure.GET("/home", handlers.HomeHandler)
	rSecure.POST("/checkmail", handlers.CheckMailHandler)

	//auth group sets the context and calls auth middleware
	rAuth := router.Group("/auth")
	rAuth.Use(middleware.ContextMiddleware(ctx), middleware.AuthMiddleware(ctx))
	rAuth.GET("/gms", handlers.MainPageHandler)

	return router
}
