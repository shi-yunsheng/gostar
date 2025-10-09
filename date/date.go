// @en Package date provides time-related parsing and formatting functionality
//
// @zh 提供时间相关的解析和格式化功能
package date

import "time"

// @en Get today's date by format
//
// @zh 根据指定格式获取今天的日期
func GetToday(format ...DateFormat) string {
	if len(format) == 0 {
		return time.Now().Format(string(FORMAT_DATETIME))
	}

	return time.Now().Format(string(format[0]))
}

// @en Get timestamp, if isMillisecond is true, return millisecond timestamp, otherwise return second timestamp
//
// @zh 获取时间戳，如果isMillisecond为true，则返回毫秒时间戳，否则返回秒时间戳
func GetTimestamp(isMillisecond ...bool) int64 {
	if len(isMillisecond) > 0 && isMillisecond[0] {
		return time.Now().UnixMilli()
	}
	return time.Now().Unix()
}
