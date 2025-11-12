package utils

import "runtime"

// 获取堆栈信息
func GetStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
