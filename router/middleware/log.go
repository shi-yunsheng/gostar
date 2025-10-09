package middleware

import (
	"fmt"
	"gostar/logger"
	"gostar/router/handler"
	"time"
)

// @en log middleware
//
// @zh 日志中间件
func LogMiddleware(next handler.Handler) handler.Handler {
	return func(w *handler.Response, r handler.Request) {
		// @en record request start time
		//
		// @zh 记录请求开始时间
		startTime := time.Now()

		// @en get request information
		//
		// @zh 获取请求信息
		method := r.Method
		path := r.URL.Path
		clientIPs := r.GetClientIP()

		// @en format client IPs as string
		// @zh 将客户端IP格式化为字符串
		var clientIP string
		if len(clientIPs) > 0 {
			clientIP = clientIPs[0] // @en use first IP as primary
			if len(clientIPs) > 1 {
				clientIP = clientIP + " (+" + fmt.Sprintf("%d", len(clientIPs)-1) + " more)"
			}
		} else {
			clientIP = "unknown"
		}

		// @en output request information
		//
		// @zh 输出请求信息
		logger.I("Request received: %s %s from IP: %s", method, path, clientIP)

		// @en continue to process request
		//
		// @zh 继续处理请求
		next(w, r)

		// @en calculate request processing time
		//
		// @zh 计算请求处理时间
		duration := time.Since(startTime)

		// @en output request completion information
		//
		// @zh 输出请求完成信息
		switch {
		case w.StatusCode >= 100 && w.StatusCode < 200:
			// @en 1xx informational response
			//
			// @zh 1xx 信息性响应
			logger.I("Request completed: %s %s - Status: %d - From IP: %s - Duration: %v",
				method, path, w.StatusCode, clientIP, duration)
		case w.StatusCode >= 200 && w.StatusCode < 300:
			// @en 2xx success
			//
			// @zh 2xx 成功
			logger.S("Request completed: %s %s - Status: %d - From IP: %s - Duration: %v",
				method, path, w.StatusCode, clientIP, duration)
		case w.StatusCode >= 300 && w.StatusCode < 400:
			// @en 3xx redirect
			//
			// @zh 3xx 重定向
			logger.I("Request completed: %s %s - Status: %d - From IP: %s - Duration: %v",
				method, path, w.StatusCode, clientIP, duration)
		case w.StatusCode >= 400 && w.StatusCode < 500:
			// @en 4xx client error
			//
			// @zh 4xx 客户端错误
			logger.W("Request completed: %s %s - Status: %d - From IP: %s - Duration: %v",
				method, path, w.StatusCode, clientIP, duration)
		case w.StatusCode >= 500:
			// @en 5xx server error
			//
			// @zh 5xx 服务器错误
			logger.E("Request completed: %s %s - Status: %d - From IP: %s - Duration: %v",
				method, path, w.StatusCode, clientIP, duration)
		default:
			// @en unknown status code
			//
			// @zh 未知状态码
			logger.W("Request completed: %s %s - Status: %d - From IP: %s - Duration: %v",
				method, path, w.StatusCode, clientIP, duration)
		}
	}
}
