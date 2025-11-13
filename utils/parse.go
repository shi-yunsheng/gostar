// 提供解析和格式化的工具函数
package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// 解析大小字符串并返回字节数
// 支持单位：B、KB、MB、GB、TB（不区分大小写）
// 例如："1KB", "2.5MB", "1GB"
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	// 使用正则表达式匹配数字和单位
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGT]?B?)$`)
	matches := re.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(sizeStr)))

	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}
	// 解析数字部分
	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil || num <= 0 {
		return 0, fmt.Errorf("invalid number: %s", matches[1])
	}
	// 解析单位部分
	unit := matches[2]
	if unit == "" || unit == "B" {
		return int64(num), nil
	}
	// 根据单位获取倍数
	var multiplier int64
	switch unit {
	case "KB", "K":
		multiplier = 1024
	case "MB", "M":
		multiplier = 1024 * 1024
	case "GB", "G":
		multiplier = 1024 * 1024 * 1024
	case "TB", "T":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unsupported unit: %s", unit)
	}

	return int64(num * float64(multiplier)), nil
}
