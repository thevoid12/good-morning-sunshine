package routes

import (
	"context"
	"gms/client/middleware"
	assests "gms/client/ui/assets"
	"gms/client/ui/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Initialize(ctx context.Context, l *zap.Logger) (router *gin.Engine) {
	l.Sugar().Info("Initializing logger")

	router = gin.Default()
	router.Use(gin.Recovery())
	//Assests and Tailwind
	router.StaticFS("/assets", http.FS(assests.AssestFS))

	//secure group
	rSecure := router.Group("/sec")

	rSecure.Use(middleware.ContextMiddleware(ctx))
	rSecure.GET("/home", handlers.HomeHandler)
	router.GET("/", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "sec/home") })
	rSecure.POST("/checkmail", handlers.CheckMailHandler)

	//auth group sets the context and calls auth middleware
	rAuth := router.Group("/auth")
	rAuth.Use(middleware.ContextMiddleware(ctx), middleware.AuthMiddleware(ctx))
	rAuth.GET("/gms", handlers.MainPageHandler)
	rAuth.POST("/gms/submit", handlers.NewMailRecordHandler)

	rAuth.POST("/gms/deactivate/:id", handlers.DeactivateRecordHandler)

	for _, route := range router.Routes() {
		l.Sugar().Infof("Route: %s %s", route.Method, route.Path)
}

	return router
}
