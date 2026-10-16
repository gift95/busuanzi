package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"busuanzi/library/jwtutil"
	"busuanzi/library/tool"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestCorsWhiteList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	viper.Set("Web.Cors", "https://a.example.com, https://b.example.com")

	r := gin.New()
	r.Use(Cors())
	r.GET("/api", func(c *gin.Context) { c.Status(204) })

	// 白名单内：允许并回显
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api", nil)
	req.Header.Set("Origin", "https://a.example.com")
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://a.example.com" {
		t.Fatalf("expected allow origin, got %q", got)
	}

	// 白名单外：不得回显（原实现存在此漏洞）
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("evil origin should not be echoed, got %q", got)
	}

	// 无 Origin（如 curl）：不设置 CORS 头
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api", nil)
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("no-origin request should not get CORS header, got %q", got)
	}
}

func TestCorsWildcard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	viper.Set("Web.Cors", "*")

	r := gin.New()
	r.Use(Cors())
	r.GET("/api", func(c *gin.Context) { c.Status(204) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api", nil)
	req.Header.Set("Origin", "https://any.example.com")
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://any.example.com" {
		t.Fatalf("wildcard should echo origin, got %q", got)
	}
}

func TestIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	viper.Set("bsz.secret", "test-secret-123456")
	viper.Set("bsz.privacy", true) // 隐私模式（默认开启）：身份为 MD5(IP+UA)

	r := gin.New()
	r.Use(Identity())
	r.GET("/api", func(c *gin.Context) {
		c.String(200, c.GetString("user_identity"))
	})

	// 无 token：生成身份并签发 Set-Bsz-Identity
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	r.ServeHTTP(w, req)
	want := tool.Md5("1.2.3.4")
	if body := w.Body.String(); body != want {
		t.Fatalf("privacy identity should be md5(ip+ua), got %q want %q", body, want)
	}
	if w.Header().Get("Set-Bsz-Identity") == "" {
		t.Fatal("expected Set-Bsz-Identity header")
	}

	// 关闭隐私模式：身份为明文 IP
	viper.Set("bsz.privacy", false)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api", nil)
	req.RemoteAddr = "5.6.7.8:1234"
	r.ServeHTTP(w, req)
	if body := w.Body.String(); body != "5.6.7.8" {
		t.Fatalf("non-privacy identity should be raw ip, got %q", body)
	}
	viper.Set("bsz.privacy", true)

	// 携带合法 token：复用身份
	token := jwtutil.Generate("user-x")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if got := w.Body.String(); got != "user-x" {
		t.Fatalf("expected user-x, got %q", got)
	}

	// 非法 token：重新签发新身份
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api", nil)
	req.Header.Set("Authorization", "Bearer user-y.bad.signature")
	r.ServeHTTP(w, req)
	if got := w.Body.String(); got == "user-y" {
		t.Fatal("forged identity should be rejected")
	}
}
