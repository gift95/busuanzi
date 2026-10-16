package jwtutil

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestGenerateAndCheck(t *testing.T) {
	viper.Set("bsz.secret", "test-secret-123456")

	token := Generate("user1")
	if got := Check(token); got != "user1" {
		t.Fatalf("expected user1, got %q", got)
	}

	// 篡改身份或签名都应校验失败
	if got := Check("user2." + token); got != "" {
		t.Fatalf("tampered identity should fail, got %q", got)
	}
	if got := Check("user1.0.abcdef"); got != "" {
		t.Fatalf("tampered signature should fail, got %q", got)
	}
	if got := Check(""); got != "" {
		t.Fatalf("empty token should fail, got %q", got)
	}
	if got := Check("a.b"); got != "" {
		t.Fatalf("malformed token should fail, got %q", got)
	}
}

func TestExpiredToken(t *testing.T) {
	viper.Set("bsz.secret", "test-secret-123456")

	token := Generate("user1")
	// 手动构造已过期 token：身份 + 过去的时间戳 + 对应签名
	expired := "user1.1." + sign("user1", 1)
	if got := Check(expired); got != "" {
		t.Fatalf("expired token should fail, got %q", got)
	}
	_ = token
	_ = time.Now
}
