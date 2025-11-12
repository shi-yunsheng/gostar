package middleware

import "github.com/shi-yunsheng/gostar/router/handler"

// 中间件函数
type Middleware func(next handler.Handler) handler.Handler
