package core

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"busuanzi/library/tool"
	"busuanzi/process/redisutil"

	"github.com/spf13/viper"
)

// 数据存储（key 前缀取配置 redis.prefix，默认 bsz）：
// sitePv  String  bsz:site_pv:{host}
// siteUv  Set     bsz:site_uv:{host}
// pagePv  ZSet    bsz:page_pv:{host}           member={path}
// pageUv  Set     bsz:page_uv:{host}:{path}
// 注：README 描述 UV 为 HyperLogLog，实际实现为 Set（SADD/SCARD）

type Counts struct {
	SitePv int64
	SiteUv int64
	PagePv int64
	PageUv int64
}

type RKeys struct {
	SitePvKey  string
	SiteUvKey  string
	PagePvKey  string
	PageUvKey  string
	SiteUnique string
	PathUnique string
}

// Count 计数并返回最新统计值，全部命令合并为一次 pipeline 往返
func Count(ctx context.Context, host string, path string, userIdentity string) Counts {
	rk := getKeys(host, path)

	pipe := redisutil.RDB.Pipeline()
	incr := pipe.Incr(ctx, rk.SitePvKey)
	zincr := pipe.ZIncrBy(ctx, rk.PagePvKey, 1, rk.PathUnique)
	pipe.SAdd(ctx, rk.SiteUvKey, userIdentity)
	pipe.SAdd(ctx, rk.PageUvKey, userIdentity)
	suv := pipe.SCard(ctx, rk.SiteUvKey)
	puv := pipe.SCard(ctx, rk.PageUvKey)
	_, err := pipe.Exec(ctx)

	go setExpire(ctx, rk.SiteUvKey, rk.PageUvKey, rk.SitePvKey, rk.PagePvKey)

	counts := Counts{}
	if err != nil {
		log.Printf("[ERROR] count: %v", err)
		return counts
	}
	counts.SitePv, _ = incr.Result()
	counts.PagePv = int64(zincr.Val())
	counts.SiteUv, _ = suv.Result()
	counts.PageUv, _ = puv.Result()
	return counts
}

// Put 仅写入数据，不返回
func Put(ctx context.Context, host string, path string, userIdentity string) {
	rk := getKeys(host, path)

	pipe := redisutil.RDB.Pipeline()
	pipe.Incr(ctx, rk.SitePvKey)
	pipe.ZIncrBy(ctx, rk.PagePvKey, 1, rk.PathUnique)
	pipe.SAdd(ctx, rk.SiteUvKey, userIdentity)
	pipe.SAdd(ctx, rk.PageUvKey, userIdentity)
	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("[ERROR] put: %v", err)
	}

	go setExpire(ctx, rk.SiteUvKey, rk.PageUvKey, rk.SitePvKey, rk.PagePvKey)
}

// Get 仅读取统计值，不计数
func Get(ctx context.Context, host string, path string) Counts {
	rk := getKeys(host, path)

	pipe := redisutil.RDB.Pipeline()
	gpv := pipe.Get(ctx, rk.SitePvKey)
	zpv := pipe.ZScore(ctx, rk.PagePvKey, rk.PathUnique)
	suv := pipe.SCard(ctx, rk.SiteUvKey)
	puv := pipe.SCard(ctx, rk.PageUvKey)
	_, err := pipe.Exec(ctx)

	go setExpire(ctx, rk.SiteUvKey, rk.PageUvKey, rk.SitePvKey, rk.PagePvKey)

	counts := Counts{}
	if err != nil {
		log.Printf("[ERROR] get: %v", err)
		return counts
	}
	if v, err := gpv.Int64(); err == nil {
		counts.SitePv = v
	}
	counts.PagePv = int64(zpv.Val())
	counts.SiteUv, _ = suv.Result()
	counts.PageUv, _ = puv.Result()
	return counts
}

// expireCache 记录每个 key 上次刷新过期时间，1 分钟内同一 key 只刷新一次，降低写放大
var expireCache sync.Map // key -> time.Time

func setExpire(ctx context.Context, key ...string) {
	ttl := viper.GetInt("bsz.expire")
	if ttl == 0 {
		return
	}

	now := time.Now()
	for _, k := range key {
		if v, ok := expireCache.Load(k); ok && now.Sub(v.(time.Time)) < time.Minute {
			continue
		}
		// 缓存条目过多时整体重建，避免无限增长
		if cacheSize() > 10000 {
			expireCache = sync.Map{}
		}
		expireCache.Store(k, now)
		redisutil.RDB.Expire(ctx, k, time.Duration(ttl)*time.Second)
	}
}

func cacheSize() int {
	n := 0
	expireCache.Range(func(_, _ any) bool { n++; return true })
	return n
}

func getKeys(host string, path string) RKeys {
	var siteUnique = host
	var pathUnique = path

	// 兼容旧版本
	if viper.GetBool("bsz.pathStyle") == false {
		pathUnique = host + "&" + path
	}

	// 隐私模式（默认开启，与上游一致）：host/path 一律 MD5 混淆
	// 关闭隐私模式（BSZ_PRIVACY=false）后使用明文 host/path
	if viper.GetBool("bsz.privacy") {
		switch viper.GetString("bsz.Encrypt") {
		case "MD516":
			siteUnique = tool.Md5(siteUnique)[8:24]
			pathUnique = tool.Md5(pathUnique)[8:24]
		default: // MD532 或未配置：完整 MD5 混淆（同上游）
			siteUnique = tool.Md5(siteUnique)
			pathUnique = tool.Md5(pathUnique)
		}
	}

	redisPrefix := viper.GetString("redis.prefix")

	siteUvKey := strings.Join([]string{redisPrefix, "site_uv", siteUnique}, ":")
	pageUvKey := strings.Join([]string{redisPrefix, "page_uv", siteUnique, pathUnique}, ":")
	sitePvKey := strings.Join([]string{redisPrefix, "site_pv", siteUnique}, ":")
	pagePvKey := strings.Join([]string{redisPrefix, "page_pv", siteUnique}, ":")

	return RKeys{
		SitePvKey:  sitePvKey,
		SiteUvKey:  siteUvKey,
		PagePvKey:  pagePvKey,
		PageUvKey:  pageUvKey,
		SiteUnique: siteUnique,
		PathUnique: pathUnique,
	}
}
