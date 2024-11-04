package jwtutil

import (
	"busuanzi/library/tool"
	"strings"

	"github.com/spf13/viper"
)

// Generate a token
// not a standard JWT, only to prevent the fake data
func Generate(userIdentity string) string {
	sign := tool.Sha256(userIdentity, viper.GetString("bsz.secret"))
	return userIdentity + "." + sign
}

func Check(token string) string {
	arr := strings.Split(token, ".")
	if len(arr) != 2 {
		return ""
	}
	if sign := tool.Sha256(arr[0], viper.GetString("bsz.secret")); sign != arr[1] {
		return ""
	}
	return arr[0]
}
