package webutil

import (
	"busuanzi/app/controller"
	"busuanzi/app/middleware"
	"busuanzi/config"
	"github.com/gin-gonic/gin"
)

func initRoute(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.Use(middleware.Identity())
		api.POST("", controller.ApiHandler)
		api.GET("", controller.GetHandler)
		api.PUT("", controller.PutHandler)
	}

	r.GET("/ping", controller.PingHandler)

	static := r.Group("/")
	{
		static.Use(middleware.Cache())
		static.GET("/", controller.Index)
		static.StaticFile("/js", config.DistPath+"/busuanzi.js")
	}
	r.NoRoute(middleware.Cache(), controller.Index)
}
