// @en Package utils provides utility functions for parsing and formatting
//
// @zh 提供解析和格式化的工具函数
package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// @en ParseSize parses a size string and returns the size in bytes
// Supports units: B, KB, MB, GB, TB (case insensitive)
// Examples: "1KB", "2.5MB", "1GB"
//
// @zh 解析大小字符串并返回字节数
// 支持单位：B、KB、MB、GB、TB（不区分大小写）
// 例如："1KB", "2.5MB", "1GB"
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// @en Use regex to match number and unit
	//
	// @zh 使用正则表达式匹配数字和单位
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGT]?B?)$`)
	matches := re.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(sizeStr)))

	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	// @en Parse the numeric part
	//
	// @zh 解析数字部分
	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil || num <= 0 {
		return 0, fmt.Errorf("invalid number: %s", matches[1])
	}

	// @en Parse the unit part
	//
	// @zh 解析单位部分
	unit := matches[2]
	if unit == "" || unit == "B" {
		return int64(num), nil
	}

	// @en Get multiplier based on unit
	//
	// @zh 根据单位获取倍数
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
