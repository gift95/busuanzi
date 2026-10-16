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
		api.Use(middleware.RateLimit(), middleware.Identity())
		api.POST("", controller.ApiHandler)
		api.GET("", controller.GetHandler)
		api.PUT("", controller.PutHandler)
	}

	// JSONP 计数端点（script 标签跨域，无需 CORS/Authorization）
	r.GET("/jsonp", middleware.RateLimit(), controller.JsonpHandler)
	r.GET("/ping", controller.PingHandler)

	static := r.Group("/")
	{
		static.Use(middleware.Cache())
		static.GET("/", controller.Index)
		static.StaticFile("/js", config.DistPath+"/busuanzi.js")
		static.StaticFile("/js/jsonp", config.DistPath+"/busuanzi.jsonp.js")
		static.StaticFile("/js/lite", config.DistPath+"/busuanzi.lite.js")
		static.StaticFile("/js/lite_pjax", config.DistPath+"/busuanzi.pjax.lite.js")
	}
	r.NoRoute(middleware.Cache(), controller.Index)
}
