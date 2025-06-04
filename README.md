# 基于 [soxft](https://github.com/soxft/busuanzi)  可选是否加密IP及路径

## 自建不蒜子

> 一个基于 Golang + Redis 的简易访问量统计系统

- 统计站点的 UV, PV
- 统计子页面的 UV, PV
- 使用 Docker 一键部署
- 兼容 Pjax 技术的网页
- 支持从原版不蒜子迁移数据

## 安装

支持多种运行方式: 源码编译运行, Docker 运行. 详见: [Install](https://github.com/soxft/busuanzi/wiki/install)

或使用docker compose 部署

```yaml
version: "3.8"

services:
  bsz:
    image: "gift95/busuanzi:latest"
    container_name: busuanzi
    ports:
    # 修改映射到宿主机的端口 host:container
      - "4080:8080"                            
    volumes:
       - /opt/bsz/bsz:/app/expose
    environment:
    # 是否开启日志
      WEB_LOG: true                          
   # 是否开启debug模式
      WEB_DEBUG: false                       
 	# 跨域访问
      WEB_CORS: "*"                            
	 # 统计数据过期时间 单位秒, 请输入整数 (无任何访问, 超过这个时间后, 统计数据将被清空, 0为不过期)
      BSZ_EXPIRE: 0                            
 	# 签名密钥 // 请设置为任意长度的随机值
      BSZ_SECRET: "BSZ"                 
	# 填写你的 API 地址 需要转译 (即 用 `\/` 替代 `/`)
      API_SERVER: https:\/\/bsz.hnlyx.top\/api  
		# redis 地址
      REDIS_ADDRESS: redis:6379     
      REDIS_PASSWORD: redis
      # 路径样式 (false: url&path, true: path) 老版本请使用 false,  true 更便于数据迁移
      BSZ_PATHSTYLE: true
      # 加密算法 (MD516 / MD532/空) 
      BSZ_ENCRYPT: 
```



## 使用方式

详见: https://bsz.hnlyx.top

## 原理

- `Busuanzi` 使用 Redis 进行数据存储与检索。Redis 作为内存数据库拥有极高的读写性能，同时其独特的`RDB`与`AOF`持久化方式，使得 Redis 的数据安全得到保障。

- UV 与 PV 数据分别采用以下方式进行存储:

| index  | 数据类型        | key                               |
|--------|-------------|-----------------------------------|
| sitePv | String      | bsz:site_pv:md5(host)/ host          |
| siteUv | HyperLogLog | bsz:site_uv:md5(host)  host        |
| pagePv | ZSet        | bsz:page_pv:md5(host) / md5(path) / (host/path) |
| pageUv | HyperLogLog | bsz:site_uv:md5(host):md5(path)  / (host:path) |
