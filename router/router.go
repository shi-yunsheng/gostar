package router

import (
	"gostar/router/handler"
	"gostar/router/middleware"
	"net/http"
)

// @en new router
//
// @zh 新建路由
func NewRouter() *Router {
	return &Router{
		mux:          http.NewServeMux(),
		routes:       make(map[string]*Route),
		sortedRoutes: make([]string, 0),
		middleware:   make([]middleware.Middleware, 0),
		secretKey:    make(map[string]string),
	}
}

// @en use route
//
// @zh 使用路由
func (r *Router) UseRoute(routes []Route) {
	r.parseRoute(routes, "")

	r.sortRoutes()

	handleFunc := r.serveHTTP

	// @en load global middleware
	// @zh 加载全局中间件
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handleFunc = r.middleware[i](handleFunc)
	}

	// @en root route handler
	// @zh 根路由处理器
	r.mux.HandleFunc("/", handler.ToHttpHandler(handleFunc))
}

// @en use middleware
//
// @zh 使用中间件
func (r *Router) UseMiddleware(middleware ...middleware.Middleware) {
	r.middleware = append(r.middleware, middleware...)
}

// @en use secret key
//
// @zh 使用认证密钥
func (r *Router) UseSecretKey(key string, value string) {
	if r.secretKey == nil {
		r.secretKey = make(map[string]string)
	}
	r.secretKey[key] = value
}
