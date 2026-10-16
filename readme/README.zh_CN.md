[![soxft/busuanzi](https://socialify.cmds.run/soxft/busuanzi/image?description=1&font=Raleway&forks=1&language=1&logo=https%3A%2F%2Fraw.githubusercontent.com%2Fsoxft%2Fbusuanzi%2Fmain%2Fdist%2Ffavicon.png&name=1&owner=1&pattern=Circuit%20Board&stargazers=1&theme=Dark&cache=43200)](https://busuanzi.9420.ltd)

# 自建不蒜子 (busuanzi)

> 一个基于 Golang + Redis 的简易访问量统计系统

- 统计站点的 UV, PV
- 统计子页面的 UV, PV
- 使用 Docker 一键部署
- 隐私模式默认开启：仅存储 MD5 混淆的 host/path 与 MD5(IP+UA) 身份（`BSZ_PRIVACY=false` 可切换明文）
- JSONP 计数端点与多前端变体（完整版 / JSONP / Lite / Lite+Pjax）
- 兼容 Pjax 技术的网页
- 支持从原版不蒜子迁移数据

> 📄 完整安装、配置、API 与前端接入文档见根目录 [README.md](../README.md)

## 安装

```bash
# 方式一：Docker Compose（推荐）
docker compose up -d

# 方式二：本地构建镜像
docker build -t gift95/busuanzi:latest .
```

> ⚠️ 服务启动会校验 `BSZ_SECRET`：等于默认值或长度不足 16 位将拒绝启动，请通过环境变量设置强随机密钥。

## 使用方式

支持多种自定义属性（`data-api` / `data-prefix` / `data-style` / `pjax`），兼容 pjax 网页，支持自定义标签前缀，详见根 [README.md](../README.md)「前端接入」章节。

## 原理

- `Busuanzi` 使用 Redis 进行数据存储与检索。Redis 作为内存数据库拥有极高的读写性能，同时其独特的`RDB`与`AOF`持久化方式，使得 Redis 的数据安全得到保障。

- UV 与 PV 数据分别采用以下方式进行存储:

| index  | 数据类型 | key                             |
|--------|----------|---------------------------------|
| sitePv | String   | bsz:site_pv:{host}              |
| siteUv | Set      | bsz:site_uv:{host}              |
| pagePv | ZSet     | bsz:page_pv:{host}（member=path）|
| pageUv | Set      | bsz:page_uv:{host}:{path}       |

> 隐私模式默认开启：上表 `{host}` / `{path}` 会做 MD5 混淆（`BSZ_ENCRYPT=MD516` 取中段 16 位，`MD532` 或空取完整 MD5）；`BSZ_PRIVACY=false` 则明文存储。

## 数据迁移

- 可使用 [busuanzi-sync](https://github.com/soxft/busuanzi-sync) 工具迁移[原版不蒜子](http://busuanzi.ibruce.info)的数据
- 老版本数据迁移请关注 `BSZ_PATHSTYLE` 配置（false 兼容旧版 `url&path` 键格式）

## 其他

Logo 由 ChatGPT 设计

## 升级建议

- 请务必在升级前备份数据 (dump.rdb)
- 新老版本数据可能并不兼容, 请注意 Release 界面的说明, 谨慎升级
- 2.5.x - 2.7.x 可以使用 [bsz-transfer](https://github.com/soxft/busuanzi-transfer) 工具进行数据迁移
