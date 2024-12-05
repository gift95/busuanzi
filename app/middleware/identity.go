package middleware

import (
	"busuanzi/library/jwtutil"
	"busuanzi/library/tool"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Identity() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Set-Bsz-Identity")

		// token
		var userIdentity string
		tokenTmp := c.Request.Header.Get("Authorization")

		if tokenTmp == "" {
			// generate jwt token
			userIdentity = getUserIdentity(c)
			setBszIdentity(c, userIdentity)
		} else {
			token := strings.Replace(tokenTmp, "Bearer ", "", -1)
			// check if token is illegal
			if userIdentity = jwtutil.Check(token); userIdentity == "" {
				// fake data, regenerate jwt token
				userIdentity = getUserIdentity(c)
				setBszIdentity(c, userIdentity)
			}
		}
		c.Set("user_identity", userIdentity)
		c.Next()
	}
}

func setBszIdentity(c *gin.Context, userIdentity string) {
	uid := jwtutil.Generate(userIdentity)
	c.Writer.Header().Set("Set-Bsz-Identity", uid)
}

func getUserIdentity(c *gin.Context) string {
	// 获取客户端 IP 和 User-Agent
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()
	// 判断加密方式是否是 MD516 或 MD532
	if isEncryptMD5(viper.GetString("Bsz.Encrypt")) {
		return tool.Md5(clientIP + userAgent)
	}

	return clientIP
}

// 判断是否是 MD5 加密方式
func isEncryptMD5(encryptType string) bool {
	return encryptType == "MD516" || encryptType == "MD532"
}
