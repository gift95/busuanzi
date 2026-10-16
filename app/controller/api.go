package controller

import (
	"context"
	"net/url"
	"strings"

	"busuanzi/core"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var defaultData = gin.H{
	"project": "https://github.com/soxft/busuanzi",
	"usage":   "https://github.com/soxft/busuanzi/wiki/usage",
}

// parseReferer 校验并解析 x-bsz-referer，返回 host 与 path。
// 配置 Bsz.AllowHosts（逗号分隔）后，仅白名单内的 host 允许计数。
func parseReferer(c *gin.Context) (host, path string, ok bool) {
	referer := c.Request.Header.Get("x-bsz-referer")
	if referer == "" {
		return "", "", false
	}
	u, err := url.Parse(referer)
	if err != nil || u.Host == "" {
		return "", "", false
	}
	if hosts := viper.GetString("bsz.allowHosts"); hosts != "" {
		allowed := false
		for _, h := range strings.Split(hosts, ",") {
			if strings.EqualFold(strings.TrimSpace(h), u.Host) {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", "", false
		}
	}
	return u.Host, u.Path, true
}

func ApiHandler(c *gin.Context) {
	host, path, ok := parseReferer(c)
	if !ok {
		c.JSON(200, gin.H{
			"success": false,
			"message": "invalid referer",
			"data":    defaultData,
		})
		return
	}

	// count
	counts := core.Count(c, host, path, c.GetString("user_identity"))

	// json
	c.JSON(200, gin.H{
		"success": true,
		"message": "ok",
		"data": gin.H{
			"site_pv": counts.SitePv,
			"site_uv": counts.SiteUv,
			"page_pv": counts.PagePv,
			"page_uv": counts.PageUv,
		},
	})
}

// PutHandler 仅提交数据, 不做返回
func PutHandler(c *gin.Context) {
	host, path, ok := parseReferer(c)
	if !ok {
		c.Status(400)
		return
	}

	// 使用独立 context，避免复用已结束的请求上下文
	go core.Put(context.Background(), host, path, c.GetString("user_identity"))
	c.Status(204)
}

// GetHandler 仅获取数据, 不增加
func GetHandler(c *gin.Context) {
	host, path, ok := parseReferer(c)
	if !ok {
		c.JSON(200, gin.H{
			"success": false,
			"message": "invalid referer",
			"data":    defaultData,
		})
		return
	}

	// count
	counts := core.Get(c, host, path)

	// json
	c.JSON(200, gin.H{
		"success": true,
		"message": "ok",
		"data": gin.H{
			"site_pv": counts.SitePv,
			"site_uv": counts.SiteUv,
			"page_pv": counts.PagePv,
			"page_uv": counts.PageUv,
		},
	})
}
