package controller

import (
	"net/http"
	"net/http/httptest"
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
