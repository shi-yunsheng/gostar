package middleware

import (
	"fmt"
	"time"

	"github.com/shi-yunsheng/gostar/logger"
	"github.com/shi-yunsheng/gostar/router/handler"
)

// 日志中间件
func LogMiddleware(next handler.Handler) handler.Handler {
	return func(w *handler.Response, r handler.Request) any {
		// 记录请求开始时间
		startTime := time.Now()
		// 获取请求信息
		method := r.Method
		path := r.URL.Path
		clientIPs := r.GetClientIP()
		// 将客户端IP格式化为字符串
		var clientIP string
		if len(clientIPs) > 0 {
			clientIP = clientIPs[0]
			if len(clientIPs) > 1 {
				clientIP = clientIP + " (+" + fmt.Sprintf("%d", len(clientIPs)-1) + " more)"
			}
		} else {
			clientIP = "unknown"
		}
		// 输出请求信息
		logger.I("Request received: %s %s from IP: %s", method, path, clientIP)
		// 继续处理请求并返回结果
		response := next(w, r)
		// 计算请求处理时间
		duration := time.Since(startTime)
		// 慢请求警告
		if duration > 3*time.Second {
			logger.W("Slow request: %s %s - Duration: %v", method, path, duration)
		}
		// 输出请求完成信息
		switch {
		case w.StatusCode >= 100 && w.StatusCode < 200:
			// 1xx 信息性响应
			logger.I("Request completed: %s %s - Status: %d - From IP: %s - Duration: %v",
				method, path, w.StatusCode, clientIP, duration)
		case w.StatusCode >= 200 && w.StatusCode < 300:
			// 2xx 成功
			logger.S("Request completed: %s %s - Status: %d - From IP: %s - Duration: %v",
				method, path, w.StatusCode, clientIP, duration)
		case w.StatusCode >= 300 && w.StatusCode < 400:
			// 3xx 重定向
			logger.I("Request completed: %s %s - Status: %d - From IP: %s - Duration: %v",
				method, path, w.StatusCode, clientIP, duration)
		case w.StatusCode >= 400 && w.StatusCode < 500:
			// 4xx 客户端错误
			logger.W("Request completed: %s %s - Status: %d - From IP: %s - Duration: %v",
				method, path, w.StatusCode, clientIP, duration)
		case w.StatusCode >= 500:
			// 5xx 服务器错误
			logger.E("Request completed: %s %s - Status: %d - From IP: %s - Duration: %v",
				method, path, w.StatusCode, clientIP, duration)
		default:
			// 未知状态码
			logger.W("Request completed: %s %s - Status: %d - From IP: %s - Duration: %v",
				method, path, w.StatusCode, clientIP, duration)
		}

		return response
	}
}
