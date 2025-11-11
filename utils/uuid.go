package utils

import "github.com/google/uuid"

// @en generate a new UUID
//
// @zh 生成一个新的UUID
func NewUUID() string {
	return uuid.New().String()
}

// @en Generate an MD5 hash
//
// @zh 进行MD5哈希
func MD5(str string) string {
	return uuid.NewMD5(uuid.Nil, []byte(str)).String()
}
