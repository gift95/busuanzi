package middleware

import (
	"net/http"
	"strings"

	"busuanzi/config"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		cors := viper.GetString("Web.Cors")
		origin := c.Request.Header.Get("Origin")

		// 仅白名单匹配时才回显 Origin，避免任意站点跨域读取计数
		var allow string
		if cors == "*" {
			allow = origin
		} else {
			for _, v := range strings.Split(cors, ",") {
				if strings.EqualFold(strings.TrimSpace(v), origin) {
					allow = origin
					break
				}
			}
		}
		if allow != "" {
			c.Header("Access-Control-Allow-Origin", allow)
		}

		c.Header("Server", "busuanzi-by-xcsoft/"+config.VERSION)
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", "GET, POST, HEAD, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "*")
			c.Header("Access-Control-Max-Age", "86400")
			if allow != "" {
				c.AbortWithStatus(204)
			} else {
				c.AbortWithStatus(403)
			}
			return
		}
	}
}
