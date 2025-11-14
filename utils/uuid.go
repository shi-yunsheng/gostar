package utils

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/google/uuid"
)

// 生成一个新的UUID
func NewUUID() string {
	return uuid.New().String()
}

// 进行MD5哈希，返回标准的32位十六进制字符串
func MD5(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}
