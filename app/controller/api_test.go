package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"busuanzi/process/redisutil"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func TestParseRefererAllowHosts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	viper.Set("bsz.allowHosts", "a.example.com, b.example.com")

	r := gin.New()
	r.GET("/api", func(c *gin.Context) {
		host, path, ok := parseReferer(c)
		if !ok {
			c.JSON(400, gin.H{"ok": false})
			return
		}
		c.JSON(200, gin.H{"ok": true, "host": host, "path": path})
	})

	// 白名单内：放行并返回 host/path
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api", nil)
	req.Header.Set("x-bsz-referer", "https://a.example.com/page")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("whitelisted host should pass, got %d", w.Code)
	}

	// 白名单外：拒绝
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api", nil)
	req.Header.Set("x-bsz-referer", "https://evil.example.com/page")
	r.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("non-whitelisted host should be rejected, got %d", w.Code)
	}
}

func TestParseRefererNoWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	viper.Set("bsz.allowHosts", "")

	r := gin.New()
	r.GET("/api", func(c *gin.Context) {
		_, _, ok := parseReferer(c)
		if !ok {
			c.JSON(400, gin.H{"ok": false})
			return
		}
		c.JSON(200, gin.H{"ok": true})
	})

	// 未配置白名单时任意合法 referer 都放行
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api", nil)
	req.Header.Set("x-bsz-referer", "https://any.example.com/page")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("no whitelist should pass all, got %d", w.Code)
	}

	// 缺失 referer 拒绝
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api", nil)
	r.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("missing referer should be rejected, got %d", w.Code)
	}
}

func TestPing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mr := miniredis.RunT(t)
	defer mr.Close()
	redisutil.RDB = redis.NewClient(&redis.Options{Addr: mr.Addr()})

	r := gin.New()
	r.GET("/ping", PingHandler)

	// Redis 正常：200
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("ping should pass with redis up, got %d", w.Code)
	}
}

func TestPingRedisDown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 指向不可达地址模拟 Redis 故障
	redisutil.RDB = redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})

	r := gin.New()
	r.GET("/ping", PingHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)
	if w.Code != 503 {
		t.Fatalf("ping should return 503 with redis down, got %d", w.Code)
	}
}

func TestJsonpHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisutil.RDB = redis.NewClient(&redis.Options{Addr: mr.Addr()})

	viper.Set("bsz.allowHosts", "")
	viper.Set("bsz.privacy", true)

	r := gin.New()
	r.GET("/jsonp", JsonpHandler)

	// 缺少 callback 参数
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jsonp", nil)
	r.ServeHTTP(w, req)
	if got := w.Body.String(); got != "try{console.error('missing callback parameter')}catch{}" {
		t.Fatalf("missing callback unexpected: %s", got)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/javascript" {
		t.Fatalf("content type unexpected: %s", ct)
	}

	// 正常调用：回调包裹 JSON 数据
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/jsonp?callback=BszCallback_123", nil)
	req.Header.Set("Referer", "https://a.example.com/page")
	req.RemoteAddr = "1.2.3.4:1234"
	r.ServeHTTP(w, req)
	body := w.Body.String()
	if !strings.Contains(body, "try{BszCallback_123(") {
		t.Fatalf("expected callback invocation, got %s", body)
	}
	if !strings.Contains(body, "site_pv") {
		t.Fatalf("expected site_pv data, got %s", body)
	}

	// 白名单外 referer 被拒绝
	viper.Set("bsz.allowHosts", "a.example.com")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/jsonp?callback=C", nil)
	req.Header.Set("Referer", "https://evil.example.com/page")
	r.ServeHTTP(w, req)
	if got := w.Body.String(); got != "try{console.error('referer not allowed')}catch{}" {
		t.Fatalf("whitelist rejection unexpected: %s", got)
	}
	viper.Set("bsz.allowHosts", "")
}
