package controller

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"busuanzi/app/middleware"
	"busuanzi/core"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var defaultData = gin.H{
	"project": "https://github.com/gift95/busuanzi",
	"usage":   "https://github.com/gift95/busuanzi",
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

// JsonpHandler JSONP 计数端点：通过 <script> 标签天然跨域取数，兼容不支持 CORS 的环境。
// 无需 Authorization 头，身份与隐私模式联动（默认 MD5(IP+UA)，关闭隐私后为明文 IP）。
func JsonpHandler(c *gin.Context) {
	c.Header("Content-Type", "application/javascript")

	callback := c.Query("callback")
	if callback == "" {
		c.String(200, "try{console.error('%s')}catch{}", "missing callback parameter")
		return
	}

	referer := c.GetHeader("Referer")
	u, err := url.Parse(referer)
	if err != nil || u.Host == "" {
		c.String(200, "try{console.error('%s')}catch{}", "invalid referer")
		return
	}

	// AllowHosts 白名单校验（与 /api 一致）
	if hosts := viper.GetString("bsz.allowHosts"); hosts != "" {
		allowed := false
		for _, h := range strings.Split(hosts, ",") {
			if strings.EqualFold(strings.TrimSpace(h), u.Host) {
				allowed = true
				break
			}
		}
		if !allowed {
			c.String(200, "try{console.error('%s')}catch{}", "referer not allowed")
			return
		}
	}

	counts := core.Count(c, u.Host, u.Path, middleware.IdentityOf(c))

	data := gin.H{
		"site_pv": counts.SitePv,
		"site_uv": counts.SiteUv,
		"page_pv": counts.PagePv,
		"page_uv": counts.PageUv,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		c.String(200, "try{console.error('%s')}catch{}", "gen json failed")
		return
	}
	c.String(200, "try{%s(%s);}catch(e){}", callback, string(jsonData))
}
