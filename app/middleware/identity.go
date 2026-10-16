package middleware

import (
	"log"
	"strings"

	"busuanzi/library/jwtutil"
	"busuanzi/library/tool"

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
				log.Printf("[WARN] invalid identity token from %s, re-issue", c.ClientIP())
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
	return IdentityOf(c)
}

// IdentityOf 计算请求的匿名身份标识：
// 隐私模式（默认开启）下取 MD5(IP+UA)，关闭隐私模式后返回明文 IP
func IdentityOf(c *gin.Context) string {
	if !viper.GetBool("bsz.privacy") {
		return c.ClientIP()
	}
	return tool.Md5(c.ClientIP() + c.Request.UserAgent())
}
