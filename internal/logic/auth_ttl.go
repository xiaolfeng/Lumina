package logic

import (
	"strconv"
	"strings"
	"time"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// TokenTTL 把设置里的秒数转成有效期。空值、非数字和非正数回退到设置目录里的默认值。
func TokenTTL(raw string, key string) time.Duration {
	fallback := settingDefaultTTL(key)
	sec, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || sec <= 0 {
		return fallback
	}
	return time.Duration(sec) * time.Second
}

func settingDefaultTTL(key string) time.Duration {
	for _, def := range bConst.SettingKeyDefs {
		if def.Key != key {
			continue
		}
		sec, err := strconv.Atoi(def.Default)
		if err != nil || sec <= 0 {
			break
		}
		return time.Duration(sec) * time.Second
	}
	return time.Hour
}
