// 提供时间相关的解析和格式化功能
package date

import (
	"time"
	_ "time/tzdata"
)

// 根据指定格式获取今天的日期
func GetToday(format ...DateFormat) string {
	if len(format) == 0 {
		return time.Now().Format(string(FORMAT_DATETIME))
	}

	return time.Now().Format(string(format[0]))
}

// 获取时间戳，如果isMillisecond为true，则返回毫秒时间戳，否则返回秒时间戳
func GetTimestamp(isMillisecond ...bool) int64 {
	if len(isMillisecond) > 0 && isMillisecond[0] {
		return time.Now().UnixMilli()
	}
	return time.Now().Unix()
}

// 获取当前时间，返回 time.Time 类型
func Now() time.Time {
	return time.Now()
}
