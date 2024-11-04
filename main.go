package main

import (
	"busuanzi/config"
	"busuanzi/process/redisutil"
	"busuanzi/process/webutil"
)

func main() {
	config.Init()
	redisutil.Init()

	webutil.Init()
}
