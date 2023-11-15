package pkg

import (
	"crypto/rand"
	"fmt"
	"io"
	"time"
)

func GenerateUUID() (string, error) {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		return "", err
	}
	uuid[8] = uuid[8]&^0xc0 | 0x80 // 设置 UUID 版本为 4
	uuid[6] = uuid[6]&^0xf0 | 0x40 // 设置 UUID 变体为 RFC4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// GetCurrentDate 返回当前日期的字符串（年月日）
func GetCurrentDate() string {
	currentTime := time.Now()
	return currentTime.Format("2006/01/02")
}
