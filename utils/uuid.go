package utils

import "github.com/google/uuid"

// @en generate a new UUID
//
// @zh 生成一个新的UUID
func NewUUID() string {
	return uuid.New().String()
}
