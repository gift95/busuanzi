package config

import (
	"flag"
	"log"
	"strings"

	"github.com/spf13/viper"
)

var (
	configPath string
	DistPath   string
	VERSION    = "2.1.1"
)

func Init() {
	// get config file path
	flag.StringVar(&configPath, "c", "config.yaml", "config path")
	flag.StringVar(&DistPath, "d", "dist", "dist path")
	flag.Parse()

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("read config file error: %s", err)
	}

	// 安全校验：拒绝弱签名密钥
	if secret := viper.GetString("bsz.secret"); secret == "bsz" || len(secret) < 16 {
		log.Fatalf("[ERROR] BSZ_SECRET 过弱，请设置至少 16 位随机密钥")
	}

	// 只打印非敏感配置，避免泄露密钥/密码
	log.Printf("[INFO] Config loaded: web=%s debug=%v expire=%d pathStyle=%v encrypt=%q",
		viper.GetString("web.address"), viper.GetBool("web.debug"),
		viper.GetInt("bsz.expire"), viper.GetBool("bsz.pathStyle"), viper.GetString("bsz.encrypt"))
}
