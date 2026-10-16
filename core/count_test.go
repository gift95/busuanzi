package core

import (
	"context"
	"testing"

	"busuanzi/library/tool"
	"busuanzi/process/redisutil"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func setup(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	mr := miniredis.RunT(t)
	redisutil.RDB = redis.NewClient(&redis.Options{Addr: mr.Addr()})

	viper.Set("redis.prefix", "bsz")
	viper.Set("bsz.expire", 0)
	viper.Set("bsz.pathStyle", true)
	viper.Set("bsz.encrypt", "")
	viper.Set("bsz.privacy", false) // 测试默认走明文 key 分支
	return mr
}

func TestCountGetPut(t *testing.T) {
	mr := setup(t)
	defer mr.Close()
	ctx := context.Background()

	// 首次访问：四项均为 1
	c1 := Count(ctx, "example.com", "/", "u1")
	if c1.SitePv != 1 || c1.SiteUv != 1 || c1.PagePv != 1 || c1.PageUv != 1 {
		t.Fatalf("first visit unexpected: %+v", c1)
	}

	// 同一用户重复访问：PV+1，UV 不变
	c2 := Count(ctx, "example.com", "/", "u1")
	if c2.SitePv != 2 || c2.SiteUv != 1 || c2.PagePv != 2 || c2.PageUv != 1 {
		t.Fatalf("repeat visit unexpected: %+v", c2)
	}

	// 新用户访问不同页面：站点 UV 增加，页面统计独立
	c3 := Count(ctx, "example.com", "/about", "u2")
	if c3.SitePv != 3 || c3.SiteUv != 2 || c3.PagePv != 1 || c3.PageUv != 1 {
		t.Fatalf("new page visit unexpected: %+v", c3)
	}

	// Get 只读不计数
	g := Get(ctx, "example.com", "/")
	if g.SitePv != 3 || g.SiteUv != 2 || g.PagePv != 2 || g.PageUv != 1 {
		t.Fatalf("get unexpected: %+v", g)
	}

	// Put 只写入，不返回
	Put(ctx, "example.com", "/", "u3")
	g = Get(ctx, "example.com", "/")
	if g.SitePv != 4 || g.SiteUv != 3 || g.PagePv != 3 || g.PageUv != 2 {
		t.Fatalf("after put unexpected: %+v", g)
	}
}

func TestGetKeys(t *testing.T) {
	setup(t)
	ctx := context.Background()
	_ = ctx

	// 默认不加密、pathStyle=true
	rk := getKeys("example.com", "/a")
	if rk.SitePvKey != "bsz:site_pv:example.com" {
		t.Fatalf("site pv key unexpected: %s", rk.SitePvKey)
	}
	if rk.PagePvKey != "bsz:page_pv:example.com" {
		t.Fatalf("page pv key unexpected: %s", rk.PagePvKey)
	}
	if rk.PageUvKey != "bsz:page_uv:example.com:/a" {
		t.Fatalf("page uv key unexpected: %s", rk.PageUvKey)
	}

	// pathStyle=false：path 键兼容旧版 host&path
	viper.Set("bsz.pathStyle", false)
	rk = getKeys("example.com", "/a")
	if rk.PathUnique != "example.com&/a" {
		t.Fatalf("legacy path unique unexpected: %s", rk.PathUnique)
	}
	viper.Set("bsz.pathStyle", true)

	// 隐私模式（默认开启，同上游）：host/path 一律 MD5 混淆
	viper.Set("bsz.privacy", true)

	// encrypt 为空时也做完整 MD5（同上游 default 分支）
	viper.Set("bsz.encrypt", "")
	rk = getKeys("example.com", "/a")
	if rk.SitePvKey != "bsz:site_pv:"+tool.Md5("example.com") {
		t.Fatalf("privacy default md5 site pv key unexpected: %s", rk.SitePvKey)
	}
	if rk.PathUnique != tool.Md5("/a") {
		t.Fatalf("privacy default md5 path unique unexpected: %s", rk.PathUnique)
	}

	// MD532：完整 MD5
	viper.Set("bsz.encrypt", "MD532")
	rk = getKeys("example.com", "/a")
	if rk.SitePvKey != "bsz:site_pv:"+tool.Md5("example.com") {
		t.Fatalf("privacy md532 site pv key unexpected: %s", rk.SitePvKey)
	}

	// MD516：中段 16 位
	viper.Set("bsz.encrypt", "MD516")
	rk = getKeys("example.com", "/a")
	if rk.SiteUnique != tool.Md5("example.com")[8:24] {
		t.Fatalf("privacy md516 site unique unexpected: %s", rk.SiteUnique)
	}

	viper.Set("bsz.privacy", false) // 复位，避免影响其他测试
	viper.Set("bsz.encrypt", "")
}
