package gostar

import (
	"github.com/shi-yunsheng/gostar/router"
	"github.com/shi-yunsheng/gostar/router/middleware"
)

// 使用路由
func (g *goStar) UseRouter(routes []router.Route) {
	g.router.UseRoute(routes)
}

// 使用路由中间件
func (g *goStar) UseMiddleware(middleware ...middleware.Middleware) {
	g.router.UseMiddleware(middleware...)
}

// 使用认证密钥
func (g *goStar) UseSecretKey(secretKey map[string]string) {
	for key, value := range secretKey {
		g.router.UseSecretKey(key, value)
	}
}
