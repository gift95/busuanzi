package controller

import (
	"context"
	"net/http"
	"time"

	"busuanzi/process/redisutil"

	"github.com/gin-gonic/gin"
)

func PingHandler(c *gin.Context) {
	// 健康检查：确认 Redis 可用，超时 2 秒
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := redisutil.RDB.Ping(ctx).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "redis unavailable: " + err.Error(),
			"data":    gin.H{},
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "pong",
		"data":    gin.H{},
	})
}
