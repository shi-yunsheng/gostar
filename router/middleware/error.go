package middleware

import (
	"fmt"

	"github.com/shi-yunsheng/gostar/logger"
	"github.com/shi-yunsheng/gostar/router/handler"
)

// 错误处理中间件
func ErrorMiddleware(next handler.Handler) handler.Handler {
	return func(w *handler.Response, r handler.Request) any {
		defer func() {
			if err := recover(); err != nil {
				logger.E("Error: %v", err)
				handler.InternalServerError(w, r, fmt.Errorf("internal server error: %v", err))
			}
		}()
		// 继续处理请求并返回结果
		return next(w, r)
	}
}
