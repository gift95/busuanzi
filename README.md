<div align="center">

[English](README.en.md) | **中文**

# 自建不蒜子 (busuanzi)

> 一个基于 **Golang + Redis** 的轻量访问量统计系统

[![license](https://img.shields.io/github/license/gift95/busuanzi)](LICENSE)
[![release](https://img.shields.io/github/v/release/gift95/busuanzi)](https://github.com/gift95/busuanzi/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/gift95/busuanzi)](https://hub.docker.com/r/gift95/busuanzi)

</div>

基于 [soxft/busuanzi](https://github.com/soxft/busuanzi) 二次开发，**可选是否加密 IP 及路径**，用于自建替代原版不蒜子服务。

## ✨ 特性

- 统计站点级 / 页面级的 UV、PV
- Redis 高性能读写，支持 RDB / AOF 持久化
- Docker 一键部署，开箱即用
- 兼容 Pjax 技术的网页（SPA 无刷新跳转自动统计）
- 支持从原版不蒜子迁移数据
- 身份令牌 HMAC-SHA256 签名，支持过期自动轮换
- 计数请求合并为 Redis Pipeline，一次网络往返完成
- 内置单元测试（miniredis），`go test` 一键验证

## 🚀 快速开始

### Docker Compose（推荐）

```yaml
version: "3.8"

services:
  redis:
    image: "redis:alpine"
    volumes:
      - ./data/redis:/data

  bsz:
    image: "gift95/busuanzi:latest"
    container_name: busuanzi
    ports:
      # 修改映射到宿主机的端口 host:container
      - "4080:8080"
    volumes:
      - ./data/bsz:/app/expose
    depends_on:
      - redis
    environment:
      # 是否开启访问日志
      WEB_LOG: true
      # 是否开启 debug 模式
      WEB_DEBUG: false
      # 跨域访问：* 或逗号分隔的域名白名单
      WEB_CORS: "*"
      # 每 IP 每秒请求数（令牌桶限流），0 关闭
      WEB_RATELIMIT: 10
      # 令牌桶容量（突发上限）
      WEB_RATELIMITBURST: 20
      # 统计数据过期时间（秒），超过该时间无访问将被清空，0 为永不过期
      BSZ_EXPIRE: 0
      # 必填：签名密钥，请使用 ≥16 位随机值，可用 openssl rand -hex 16 生成
      BSZ_SECRET: "REPLACE_WITH_RANDOM_SECRET"
      # 前端脚本中的 API 地址（需转译，用 \/ 替代 /）
      API_SERVER: https:\/\/bsz.yxliu.cc
      # Redis 地址 / 密码
      REDIS_ADDRESS: redis:6379
      REDIS_PASSWORD:
      # 路径样式（true: path；false: url&path，老版本迁移请用 false）
      BSZ_PATHSTYLE: true
      # 加密算法：MD516 / MD532 / 空
      BSZ_ENCRYPT:
      # 允许统计的域名白名单（逗号分隔），空则不限制
      BSZ_ALLOWHOSTS: ""
```

> ⚠️ 服务启动时会校验 `BSZ_SECRET`：等于默认值或长度不足 16 位将**拒绝启动**。

### 源码编译

```bash
go build -o busuanzi main.go
./busuanzi -c config.yaml -d dist
```

### 环境变量

所有配置均可通过环境变量覆盖（`. ` 映射为 `_`）：

| 变量 | 对应配置 | 说明 | 默认 |
|---|---|---|---|
| WEB_ADDRESS | Web.Address | 监听地址 | 0.0.0.0:8080 |
| WEB_CORS | Web.Cors | 跨域白名单：`*` 或逗号分隔域名 | * |
| WEB_DEBUG | Web.Debug | debug 模式 | false |
| WEB_LOG | Web.Log | 访问日志 | false |
| WEB_RATELIMIT | Web.RateLimit | 每 IP 每秒请求数（令牌桶，0 关闭） | 10 |
| WEB_RATELIMITBURST | Web.RateLimitBurst | 令牌桶容量（突发上限） | 20 |
| REDIS_ADDRESS | Redis.Address | Redis 地址 | redis:6379 |
| REDIS_PASSWORD | Redis.Password | Redis 密码 | 空 |
| REDIS_DATABASE | Redis.Database | Redis 库号 | 0 |
| REDIS_TLS | Redis.TLS | 使用 TLS 连接 Redis | false |
| REDIS_PREFIX | Redis.Prefix | Redis key 前缀 | bsz |
| BSZ_EXPIRE | Bsz.Expire | 统计过期秒数，0 不过期 | 0 |
| BSZ_SECRET | Bsz.Secret | 身份令牌签名密钥（必填，≥16 位） | bsz |
| BSZ_ENCRYPT | Bsz.Encrypt | key/身份加密算法（MD516/MD532/空） | 空 |
| BSZ_PATHSTYLE | Bsz.PathStyle | 路径存储样式 | true |
| BSZ_ALLOWHOSTS | Bsz.AllowHosts | 允许统计的域名白名单（逗号分隔，空则不限制） | 空 |

## 📖 前端接入

在页面引入 `busuanzi.js`，并放置对应 ID 的元素：

```html
<script src="https://your-domain.com/js" data-api="https://your-domain.com/api"></script>

本站总访问量 <span id="busuanzi_site_pv"></span> 次
本站总访客数 <span id="busuanzi_site_uv"></span> 人
本页访问量 <span id="busuanzi_page_pv"></span> 次
本页访客数 <span id="busuanzi_page_uv"></span> 人
```

脚本支持的属性：

| 属性 | 说明 | 默认 |
|---|---|---|
| data-api | 后端 API 地址 | http://127.0.0.1:8080/api |
| data-prefix | 元素 ID 前缀 | busuanzi |
| data-style | 数字格式：default / comma / short（K/M/B） | default |
| pjax | 启用 Pjax 兼容（pushState / popstate 时自动刷新） | 无 |

## 🔌 API

所有接口需携带请求头 `x-bsz-referer: <页面完整URL>`；身份令牌放 `Authorization: Bearer <token>`（首次访问由 `Set-Bsz-Identity` 响应头下发）。

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | /api | 计数并返回最新统计 |
| GET | /api | 只读查询，不计数 |
| PUT | /api | 只计数，返回 204 |
| GET | /ping | 健康检查（含 Redis 连通性，异常返回 503） |

> 超过限流阈值（`WEB_RATELIMIT`）时 API 返回 `429`。

响应示例：

```json
{
  "success": true,
  "message": "ok",
  "data": { "site_pv": 1024, "site_uv": 42, "page_pv": 2048, "page_uv": 88 }
}
```

## 🗄️ 数据存储结构

Redis 按以下结构存储（key 前缀默认 `bsz`；开启 `BSZ_ENCRYPT=MD516/MD532` 后 host/path 会做 MD5 混淆）：

| 指标 | 类型 | Key |
|---|---|---|
| sitePv | String | bsz:site_pv:{host} |
| siteUv | Set | bsz:site_uv:{host} |
| pagePv | ZSet | bsz:page_pv:{host}（member=path） |
| pageUv | Set | bsz:page_uv:{host}:{path} |

## 🛡️ 安全建议

- `BSZ_SECRET` 必须设置为随机强密钥（否则服务拒绝启动）
- 建议开启 `BSZ_ENCRYPT=MD532`，避免 Redis key 泄露站点路径与访问者信息
- 设置 `BSZ_ALLOWHOSTS` 域名白名单，拒绝伪造 referer 刷任意站点数据
- 保持 `WEB_RATELIMIT` 开启，限制单 IP 请求频率防止刷量
- 为 Redis 设置密码，并限制可访问来源
- 仅暴露 API 端口，通过反向代理（Nginx / Caddy）提供 HTTPS

## 🔄 数据迁移

- 可使用 [busuanzi-sync](https://github.com/soxft/busuanzi-sync) 从[原版不蒜子](http://busuanzi.ibruce.info)迁移数据
- 老版本数据迁移请关注 `BSZ_PATHSTYLE`：false 使用 `url&path` 键格式，兼容旧版数据

## 🧪 开发

```bash
go build ./...   # 编译
go vet ./...     # 静态检查
go test ./...    # 单元测试
```

## 📦 升级建议

- 升级前务必备份数据（Redis dump.rdb）
- 新老版本数据可能不兼容，请留意 Release 说明
- 2.5.x - 2.7.x 可使用 [bsz-transfer](https://github.com/soxft/busuanzi-transfer) 迁移数据

## 📄 许可证

[GPL-3.0](LICENSE)

---

<picture>
  <source
    media="(prefers-color-scheme: dark)"
    srcset="
      https://api.star-history.com/svg?repos=gift95/busuanzi&type=Date&theme=dark
    "
  />
  <source
    media="(prefers-color-scheme: light)"
    srcset="
      https://api.star-history.com/svg?repos=gift95/busuanzi&type=Date
    "
  />
  <img
    alt="Star History Chart"
    src="https://api.star-history.com/svg?repos=gift95/busuanzi&type=Date"
  />
</picture>
