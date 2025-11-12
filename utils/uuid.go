package utils

import "github.com/google/uuid"

// 生成一个新的UUID
func NewUUID() string {
	return uuid.New().String()
}

// 进行MD5哈希
func MD5(str string) string {
	return uuid.NewMD5(uuid.Nil, []byte(str)).String()
}
