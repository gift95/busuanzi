package jwtutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// tokenTTL 身份令牌有效期
const tokenTTL = 30 * 24 * time.Hour

// Generate 签发身份令牌，格式：userIdentity.exp.signature（HMAC-SHA256）
// 非标准 JWT，仅用于防止伪造数据
func Generate(userIdentity string) string {
	exp := time.Now().Add(tokenTTL).Unix()
	sign := sign(userIdentity, exp)
	return fmt.Sprintf("%s.%d.%s", userIdentity, exp, sign)
}

// Check 校验令牌并返回身份，非法或已过期返回空串
func Check(token string) string {
	arr := strings.Split(token, ".")
	if len(arr) != 3 {
		return ""
	}
	exp, err := strconv.ParseInt(arr[1], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return ""
	}
	if sign(arr[0], exp) != arr[2] {
		return ""
	}
	return arr[0]
}

func sign(userIdentity string, exp int64) string {
	mac := hmac.New(sha256.New, []byte(viper.GetString("bsz.secret")))
	mac.Write([]byte(userIdentity))
	mac.Write([]byte(strconv.FormatInt(exp, 10)))
	return hex.EncodeToString(mac.Sum(nil))
}
