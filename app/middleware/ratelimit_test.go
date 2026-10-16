package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	viper.Set("web.ratelimit", 1)
	viper.Set("web.ratelimitburst", 1)

	r := gin.New()
	r.Use(RateLimit())
	r.GET("/api", func(c *gin.Context) { c.Status(204) })

	// 同一 IP 连续请求：桶容量 1，第 2 次应被限流
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api", nil)
	req.RemoteAddr = "1.2.3.4:1"
	r.ServeHTTP(w, req)
	if w.Code != 204 {
		t.Fatalf("first request should pass, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api", nil)
	req.RemoteAddr = "1.2.3.4:1"
	r.ServeHTTP(w, req)
	if w.Code != 429 {
		t.Fatalf("second request should be limited, got %d", w.Code)
	}

	// 不同 IP 不受影响
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api", nil)
	req.RemoteAddr = "5.6.7.8:1"
	r.ServeHTTP(w, req)
	if w.Code != 204 {
		t.Fatalf("other ip should pass, got %d", w.Code)
	}
}

func TestRateLimitOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	viper.Set("web.ratelimit", 1)
	viper.Set("web.ratelimitburst", 1)

	r := gin.New()
	r.Use(RateLimit())
	r.OPTIONS("/api", func(c *gin.Context) { c.Status(204) })

	// 预检请求不受限流影响
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/api", nil)
		req.RemoteAddr = "1.2.3.4:1"
		r.ServeHTTP(w, req)
		if w.Code != 204 {
			t.Fatalf("options should not be limited, got %d", w.Code)
		}
	}
}

func TestRateLimitDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	viper.Set("web.ratelimit", 0)

	r := gin.New()
	r.Use(RateLimit())
	r.GET("/api", func(c *gin.Context) { c.Status(204) })

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api", nil)
		req.RemoteAddr = "1.2.3.4:1"
		r.ServeHTTP(w, req)
		if w.Code != 204 {
			t.Fatalf("disabled rate limit should pass, got %d", w.Code)
		}
	}
}
