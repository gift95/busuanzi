<div align="center">

[**English**](README.en.md) | [中文](README.md)

# busuanzi (Self-hosted Page Counter)

> A lightweight page view counter built with **Golang + Redis**

[![license](https://img.shields.io/github/license/gift95/busuanzi)](LICENSE)
[![release](https://img.shields.io/github/v/release/gift95/busuanzi)](https://github.com/gift95/busuanzi/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/gift95/busuanzi)](https://hub.docker.com/r/gift95/busuanzi)

</div>

A fork of [soxft/busuanzi](https://github.com/soxft/busuanzi) with **privacy mode enabled by default** (MD5-masked host/path, identity = MD5(IP+UA)), toggleable via one switch, built to self-host your own busuanzi-style counter.

## ✨ Features

- Site-level and page-level UV / PV counting
- Redis-backed high performance, with RDB / AOF persistence
- One-click Docker deployment
- Pjax-compatible (automatic counting on SPA pushState / popstate navigation)
- Data migration from the original busuanzi service
- **Privacy mode on by default**: host/path always MD5-masked, identity = MD5(IP+UA); set `BSZ_PRIVACY=false` for plaintext
- **JSONP counting endpoint** (`/jsonp`): cross-origin via `<script>` tags, works where CORS is unavailable
- **Multiple front-end variants**: full / JSONP / Lite / Lite+Pjax (`/js`, `/js/jsonp`, `/js/lite`, `/js/lite_pjax`)
- HMAC-SHA256 signed identity tokens with automatic expiry rotation
- Count requests merged into a single Redis pipeline (one network round-trip)
- Built-in unit tests (miniredis), verified via `go test`

## 🚀 Quick Start

### Docker Compose (recommended)

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
      # Map the host port to the container port: host:container
      - "4080:8080"
    volumes:
      - ./data/bsz:/app/expose
    depends_on:
      - redis
    environment:
      # Enable access logs
      WEB_LOG: true
      # Enable debug mode
      WEB_DEBUG: false
      # CORS: * or comma-separated origin whitelist
      WEB_CORS: "*"
      # Requests per second per IP (token bucket rate limit), 0 to disable
      WEB_RATELIMIT: 10
      # Token bucket capacity (burst limit)
      WEB_RATELIMITBURST: 20
      # Stats expiry in seconds; counters are cleared after this idle period, 0 = never expire
      BSZ_EXPIRE: 0
      # REQUIRED: signing secret, use a random value of ≥16 chars, e.g. openssl rand -hex 16
      BSZ_SECRET: "REPLACE_WITH_RANDOM_SECRET"
      # API address baked into the front-end script (escape / as \/)
      API_SERVER: https:\/\/bsz.yxliu.cc
      # Redis address / password
      REDIS_ADDRESS: redis:6379
      REDIS_PASSWORD:
      # Path style (true: path; false: url&path, use false when migrating legacy data)
      BSZ_PATHSTYLE: true
      # Encryption algorithm: MD516 / MD532 / empty (privacy mode: empty = full MD5)
      BSZ_ENCRYPT:
      # Privacy mode (default on, same as upstream): host/path always MD5-masked, identity = MD5(IP+UA); false = plaintext host/path + raw IP
      BSZ_PRIVACY: "true"
      # Host whitelist for counting (comma-separated), empty = no restriction
      BSZ_ALLOWHOSTS: ""
```

> ⚠️ On startup the service validates `BSZ_SECRET`: it **refuses to start** if the secret equals the default or is shorter than 16 characters.

### Build from source

```bash
go build -o busuanzi main.go
./busuanzi -c config.yaml -d dist
```

### Environment variables

Every setting can be overridden via environment variables (`.` maps to `_`):

| Variable | Config key | Description | Default |
|---|---|---|---|
| WEB_ADDRESS | Web.Address | Listen address | 0.0.0.0:8080 |
| WEB_CORS | Web.Cors | CORS whitelist: `*` or comma-separated origins | * |
| WEB_DEBUG | Web.Debug | Debug mode | false |
| WEB_LOG | Web.Log | Access log | false |
| WEB_RATELIMIT | Web.RateLimit | Requests per second per IP (token bucket, 0 = off) | 10 |
| WEB_RATELIMITBURST | Web.RateLimitBurst | Token bucket capacity (burst) | 20 |
| REDIS_ADDRESS | Redis.Address | Redis address | redis:6379 |
| REDIS_PASSWORD | Redis.Password | Redis password | empty |
| REDIS_DATABASE | Redis.Database | Redis DB number | 0 |
| REDIS_TLS | Redis.TLS | Use TLS for Redis | false |
| REDIS_PREFIX | Redis.Prefix | Redis key prefix | bsz |
| BSZ_EXPIRE | Bsz.Expire | Stats expiry in seconds, 0 = never | 0 |
| BSZ_SECRET | Bsz.Secret | Identity token signing secret (required, ≥16 chars) | bsz |
| BSZ_ENCRYPT | Bsz.Encrypt | Key/identity masking in privacy mode (MD516/MD532, empty = full MD5) | empty (full MD5) |
| BSZ_PRIVACY | Bsz.Privacy | Privacy mode: MD5-masked host/path + identity MD5(IP+UA); false = plaintext host/path + raw IP | true |
| BSZ_PATHSTYLE | Bsz.PathStyle | Path storage style | true |
| BSZ_ALLOWHOSTS | Bsz.AllowHosts | Host whitelist for counting (comma-separated, empty = unrestricted) | empty |

## 📖 Front-end Integration

Load `busuanzi.js` on your page and place elements with the matching IDs:

```html
<script src="https://your-domain.com/js" data-api="https://your-domain.com/api"></script>

Site total views <span id="busuanzi_site_pv"></span>
Site total visitors <span id="busuanzi_site_uv"></span>
This page views <span id="busuanzi_page_pv"></span>
This page visitors <span id="busuanzi_page_uv"></span>
```

Script attributes:

| Attribute | Description | Default |
|---|---|---|
| data-api | Backend API address | http://127.0.0.1:8080/api |
| data-prefix | Element ID prefix | busuanzi |
| data-style | Number format: default / comma / short (K/M/B) | default |
| pjax | Enable Pjax compatibility (auto refresh on pushState / popstate) | none |

**Front-end script variants** (all bundled in the Docker image):

| Script | Notes |
|---|---|
| `/js` (busuanzi.js) | Full version: supports `data-api` / `data-prefix` / `data-style` / `pjax` |
| `/js/jsonp` (busuanzi.jsonp.js) | JSONP version: `<script>` tag cross-origin, no CORS needed, works in legacy environments |
| `/js/lite` (busuanzi.lite.js) | Lite version: only `prefix`, no attribute config or number formatting, smallest footprint |
| `/js/lite_pjax` (busuanzi.pjax.lite.js) | Lite + Pjax auto-refresh |

JSONP usage (`data-api` points to `/jsonp`):

```html
<script src="https://your-domain.com/js/jsonp" data-api="https://your-domain.com/jsonp"></script>
```

## 🔌 API

All endpoints require the `x-bsz-referer: <full page URL>` header; the identity token goes in `Authorization: Bearer <token>` (issued via the `Set-Bsz-Identity` response header on first visit).

| Method | Path | Description |
|---|---|---|
| POST | /api | Count and return latest stats |
| GET | /api | Read-only query, does not count |
| PUT | /api | Count only, returns 204 |
| GET | /jsonp?callback=xxx | JSONP counting (`Referer` header auth, returns `try{callback(json)}catch(e){}`) |
| GET | /ping | Health check (includes Redis connectivity, 503 when down) |

> API returns `429` when the rate limit (`WEB_RATELIMIT`) is exceeded.
> `/jsonp` needs no `Authorization` header: identity is computed server-side per privacy mode (default `MD5(IP+UA)`).

Example response:

```json
{
  "success": true,
  "message": "ok",
  "data": { "site_pv": 1024, "site_uv": 42, "page_pv": 2048, "page_uv": 88 }
}
```

## 🗄️ Data Storage Structure

Data is stored in Redis as follows (key prefix defaults to `bsz`):

| Metric | Type | Key |
|---|---|---|
| sitePv | String | bsz:site_pv:{host} |
| siteUv | Set | bsz:site_uv:{host} |
| pagePv | ZSet | bsz:page_pv:{host} (member=path) |
| pageUv | Set | bsz:page_uv:{host}:{path} |

> **Privacy mode is on by default** (`BSZ_PRIVACY=true`, same as upstream): `{host}` / `{path}` above are MD5-masked
> (`BSZ_ENCRYPT=MD516` keeps 16 middle chars, `MD532` or empty keeps the full MD5), so Redis keys don't leak site paths;
> set `BSZ_PRIVACY=false` for plaintext host/path (handy for debugging and migration).

## 🛡️ Security Tips

- `BSZ_SECRET` must be a random strong secret (the service refuses to start otherwise)
- **Keep `BSZ_PRIVACY=true` (default)**: Redis keys don't leak site paths or visitor IPs
- Optionally use `BSZ_ENCRYPT=MD516` (shorter keys, slightly weaker masking) or `MD532`
- Set `BSZ_ALLOWHOSTS` to block forged referers from inflating arbitrary sites
- Keep `WEB_RATELIMIT` on to prevent abuse from a single IP
- Password-protect Redis and restrict which hosts can reach it
- Expose only the API port and serve HTTPS via a reverse proxy (Nginx / Caddy)

## 🔄 Data Migration

- Use [busuanzi-sync](https://github.com/soxft/busuanzi-sync) to migrate data from the [original busuanzi](http://busuanzi.ibruce.info)
- For legacy data migration watch `BSZ_PATHSTYLE`: false keeps the `url&path` key format compatible with old data

## 🧪 Development

```bash
go build ./...   # compile
go vet ./...     # static analysis
go test ./...    # unit tests
```

## 📦 Upgrade Notes

- Always back up your data (Redis dump.rdb) before upgrading
- New versions may not be data-compatible; check the Release notes
- **Since v2.2.0 privacy mode is on by default**: if you upgrade from an older build that used plaintext keys, set `BSZ_PRIVACY=false` to keep reading existing data, or accept a fresh count
- 2.5.x - 2.7.x users can migrate with [bsz-transfer](https://github.com/soxft/busuanzi-transfer)

## 📄 License

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
