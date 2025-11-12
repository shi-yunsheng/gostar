package logger

import (
	"fmt"

	"github.com/shi-yunsheng/gostar/date"
)

const (
	logReset  = "\033[0m"        // reset
	logBlue   = "\033[94m"       // blue
	logYellow = "\033[33m"       // yellow
	logOrange = "\033[38;5;208m" // orange
	logGreen  = "\033[32m"       // green
	logRed    = "\033[31m"       // red
	logNormal = "\033[0m"        // normal
)

// 获取日志级别颜色
func getLogLevelColor(level string) string {
	switch level {
	case "INFO":
		return logBlue
	case "WARN":
		return logYellow
	case "ERROR":
		return logRed
	case "SUCCESS":
		return logGreen
	case "DEBUG":
		return logOrange
	default:
		return logNormal
	}
}

// 获取日志级别消息
func getLogLevelMessage(level string, message string) string {
	color := getLogLevelColor(level)
	// 根据日期格式获取日期时间
	datetime := date.GetToday(dateFormat)
	// 使用日志格式
	message = fmt.Sprintf(logFormat, datetime, level[0:1], message)

	return color + message + logReset
}
