package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

// visitor 记录单个 IP 的限流器与最后活跃时间
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimit 按客户端 IP 令牌桶限流，防止刷量。
// Web.RateLimit 为每秒许可数（0 关闭），Web.RateLimitBurst 为桶容量（默认 2 倍每秒许可数）。
func RateLimit() gin.HandlerFunc {
	limit := viper.GetFloat64("web.ratelimit")
	if limit <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	burst := viper.GetInt("web.ratelimitburst")
	if burst <= 0 {
		burst = int(limit) * 2
	}

	var mu sync.Mutex
	visitors := make(map[string]*visitor)

	// 定期清理超过 5 分钟未活跃的条目，避免 map 无限增长
	go func() {
		for range time.Tick(time.Minute) {
			mu.Lock()
			now := time.Now()
			for ip, v := range visitors {
				if now.Sub(v.lastSeen) > 5*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		// 预检请求不受限流影响
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		ip := c.ClientIP()
		mu.Lock()
		v, ok := visitors[ip]
		if !ok {
			v = &visitor{limiter: rate.NewLimiter(rate.Limit(limit), burst)}
			visitors[ip] = v
		}
		v.lastSeen = time.Now()
		mu.Unlock()

		if !v.limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "too many requests",
				"data":    gin.H{},
			})
			return
		}
		c.Next()
	}
}
