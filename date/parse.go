package date

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// 解析时间持续字符串并返回 time.Duration
// 支持多种时间单位：ns, us, ms, s, m, h, d, M, y
// 例如："1h30m", "2.5d", "1y6M"
func ParseTimeDuration(duration string) (time.Duration, error) {
	// 定义单位到秒的转换因子
	unitMap := map[string]float64{
		"ns": 1e-9,        // nanoseconds
		"us": 1e-6,        // microseconds
		"ms": 1e-3,        // milliseconds
		"s":  1,           // seconds
		"m":  60,          // minutes
		"h":  3600,        // hours
		"d":  86400,       // days (24 hours)
		"M":  30 * 86400,  // months (calculated as 30 days)
		"y":  365 * 86400, // years (calculated as 365 days)
	}

	var totalSeconds float64 // total seconds
	index := 0               // current parsing position
	length := len(duration)  // string length
	// 遍历字符串解析每个时间单位
	for index < length {
		// 跳过空白字符
		if duration[index] == ' ' {
			index++
			continue
		}
		// 解析数字部分（支持整数和小数）
		start := index
		hasDecimal := false
		for index < length {
			c := duration[index]
			if c >= '0' && c <= '9' {
				index++
				continue
			}
			if c == '.' {
				if hasDecimal {
					return 0, fmt.Errorf("invalid number format: multiple decimal points")
				}
				hasDecimal = true
				index++
				continue
			}
			break
		}
		// 检查是否找到数字
		if start == index {
			return 0, fmt.Errorf("missing number at position %d", index)
		}
		// 提取数字字符串并转换为浮点数
		numStr := duration[start:index]
		value, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number '%s': %w", numStr, err)
		}
		// 解析单位部分
		unitStart := index
		for index < length {
			c := duration[index]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				index++
			} else {
				break
			}
		}
		unit := duration[unitStart:index]
		if unit == "" {
			return 0, fmt.Errorf("missing unit after number '%s'", numStr)
		}
		// 获取单位对应的秒数因子
		factor, exists := unitMap[unit]
		if !exists {
			// 尝试匹配单位（区分大小写）
			validUnits := []string{"y", "M", "d", "h", "m", "s", "ms", "us", "ns"}
			return 0, fmt.Errorf("unknown unit '%s', valid units: %v", unit, validUnits)
		}
		// 计算秒数值
		totalSeconds += value * factor
		// 检查是否超出浮点数表示范围
		if math.IsInf(totalSeconds, 0) {
			return 0, fmt.Errorf("duration value too long")
		}
	}

	return time.ParseDuration(fmt.Sprintf("%vs", totalSeconds))
}

// 解析以"-"或"/"分割的日期字符串
// 支持多种日期格式，可包含时间
// 例如："2024-06-01", "2024/6/1", "2024-06-01 12:30:45"
func ParseTimeString(date string) (time.Time, error) {
	// 定义支持的日期格式
	layouts := []string{
		"2006-01-02",
		"2006/01/02",
		"2006-1-2",
		"2006/1/2",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-1-2 15:04:05",
		"2006/1/2 15:04:05",
	}
	// 依次尝试每种格式
	for _, layout := range layouts {
		t, err := time.Parse(layout, date)
		if err == nil {
			return t, nil
		}
	}
	// 如果都无法解析，返回错误
	return time.Time{}, fmt.Errorf("无法解析日期字符串: %s", date)
}

// 解析时间戳字符串并返回 time.Time
// 支持Unix时间戳（秒）、毫秒时间戳、微秒时间戳和纳秒时间戳
// 例如："1640995200", "1640995200000", "1640995200000000", "1640995200000000000"
func ParseTimestamp(timestamp string) (time.Time, error) {
	// 将时间戳解析为整数
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("无法解析时间戳: %s", timestamp)
	}
	// 根据长度确定时间戳精度
	length := len(timestamp)
	switch {
	case length <= 10:
		// Unix时间戳（秒）
		return time.Unix(ts, 0), nil
	case length <= 13:
		// 毫秒时间戳
		return time.Unix(ts/1000, (ts%1000)*1e6), nil
	case length <= 16:
		// 微秒时间戳
		return time.Unix(ts/1e6, (ts%1e6)*1e3), nil
	case length <= 19:
		// 纳秒时间戳
		return time.Unix(ts/1e9, ts%1e9), nil
	default:
		return time.Time{}, fmt.Errorf("时间戳长度过长: %s", timestamp)
	}
}
