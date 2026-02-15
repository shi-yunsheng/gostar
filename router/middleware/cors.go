package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/shi-yunsheng/gostar/router/handler"
)

// CORS中间件
func CORSMiddleware(allowedOrigins []string) Middleware {
	defaultAllowedHeaders := []string{
		"Content-Type",
		"Authorization",
		"Accept",
		"X-Requested-With",
	}

	return func(next handler.Handler) handler.Handler {
		return func(w *handler.Response, r handler.Request) any {
			origin := r.GetHeader("Origin")

			// 如果没有 Origin 头，直接放行
			if origin == "" {
				return next(w, r)
			}

			// 检查是否允许该来源
			allowed := false
			if len(allowedOrigins) == 0 || allowedOrigins[0] == "*" {
				allowed = true
			} else {
				if slices.Contains(allowedOrigins, origin) {
					allowed = true
				}
			}

			// 设置跨域响应头
			if allowed {
				w.SetHeader("Access-Control-Allow-Origin", origin)
				w.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

				// 处理 OPTIONS 请求时，从请求头中读取客户端想发送的 headers
				if strings.ToUpper(r.Method) == "OPTIONS" {
					// 从 Access-Control-Request-Headers 读取
					requestHeaders := r.GetHeader("Access-Control-Request-Headers")
					if requestHeaders != "" {
						// 直接返回客户端询问的 headers
						w.SetHeader("Access-Control-Allow-Headers", requestHeaders)
					} else {
						// 如果没有询问，返回默认允许的 headers
						w.SetHeader("Access-Control-Allow-Headers", strings.Join(defaultAllowedHeaders, ", "))
					}
				} else {
					// 非 OPTIONS 请求，返回默认允许的 headers
					w.SetHeader("Access-Control-Allow-Headers", strings.Join(defaultAllowedHeaders, ", "))
				}

				w.SetHeader("Access-Control-Allow-Credentials", "true")
				w.SetHeader("Access-Control-Max-Age", "86400") // 24小时
				w.SetHeader("Vary", "Origin")
			}

			// 处理 OPTIONS 请求
			if strings.ToUpper(r.Method) == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return nil
			}

			return next(w, r)
		}
	}
}
