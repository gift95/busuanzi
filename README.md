[![soxft/busuanzi](https://socialify.cmds.run/soxft/busuanzi/image?description=1&font=Raleway&forks=1&language=1&logo=https%3A%2F%2Fraw.githubusercontent.com%2Fsoxft%2Fbusuanzi%2Fmain%2Fdist%2Ffavicon.png&name=1&owner=1&pattern=Circuit%20Board&stargazers=1&theme=Dark&cache=43200)](https://busuanzi.9420.ltd)

# 基于 soxft 改动 可选是否加密IP及路径

## 自建不蒜子

> 一个基于 Golang + Redis 的简易访问量统计系统

- 统计站点的 UV, PV
- 统计子页面的 UV, PV
- 使用 Docker 一键部署
- 隐私保障 仅存储 HASH(可选)
- 兼容 Pjax 技术的网页
- 支持从原版不蒜子迁移数据

## 安装

支持多种运行方式: 源码编译运行, Docker 运行. 详见: [Install](https://github.com/soxft/busuanzi/wiki/install)
docker pull gift95/busuanzi:latest

## 使用方式

支持多种自定义属性, 兼容 pjax 网页, 支持自定义 标签前缀. 详见: [使用文档](https://github.com/soxft/busuanzi/wiki/usage)

## 原理

- `Busuanzi` 使用 Redis 进行数据存储与检索。Redis 作为内存数据库拥有极高的读写性能，同时其独特的`RDB`与`AOF`持久化方式，使得 Redis 的数据安全得到保障。

- UV 与 PV 数据分别采用以下方式进行存储:

| index  | 数据类型        | key                               |
|--------|-------------|-----------------------------------|
| sitePv | String      | bsz:site_pv:md5(host)/ host          |
| siteUv | HyperLogLog | bsz:site_uv:md5(host)  host        |
| pagePv | ZSet        | bsz:page_pv:md5(host) / md5(path) / (host/path) |
| pageUv | HyperLogLog | bsz:site_uv:md5(host):md5(path)  / (host:path) |
