package middleware

import "gostar/router/handler"

// @en middleware function
//
// @zh 中间件函数
type Middleware func(next handler.Handler) handler.Handler
