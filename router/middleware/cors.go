package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/shi-yunsheng/gostar/router/handler"
)

// CORS中间件
func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(next handler.Handler) handler.Handler {
		return func(w *handler.Response, r handler.Request) any {
			origin := r.GetHeader("Origin")
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
				w.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.SetHeader("Access-Control-Allow-Credentials", "true")
				// 24小时
				w.SetHeader("Access-Control-Max-Age", "86400")
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
