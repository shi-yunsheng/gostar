package gostar

import (
	"net/http"

	"github.com/shi-yunsheng/gostar/logger"
	"github.com/shi-yunsheng/gostar/model"
	"github.com/shi-yunsheng/gostar/router"
	"github.com/shi-yunsheng/gostar/router/handler"
	"github.com/shi-yunsheng/gostar/router/middleware"
)

var (
	instance *goStar
)

type goStar struct {
	version string
	config  *config
	server  *http.Server
	router  *router.Router
}

// 新建GoStar实例
func New(configName ...string) *goStar {
	instance = &goStar{
		version: "1.0.31-beta",
		config:  getConfig(configName...),
		router:  router.NewRouter(),
	}
	instance.initGoStar()
	return instance
}

// 初始化GoStar
func (g *goStar) initGoStar() {
	// 如果调试模式已开启，则开启各个组件的调试模式
	if g.config.Debug {
		middleware.EnableDebug()
		handler.EnableDebug()
	}

	g.initDate()
	g.initLog()
	model.InitDB(g.config.Database)
	model.InitRedis(g.config.Redis)
	// 使用默认路由中间件
	g.router.UseMiddleware(
		middleware.ErrorMiddleware,
		middleware.LogMiddleware,
		middleware.CORSMiddleware(g.config.AllowedOrigins),
	)
}

// 返回GoStar的版本
func (g *goStar) Version() string {
	return g.version
}

// 运行GoStar
func (g *goStar) Run() error {
	logger.I("GoStar is running on " + g.config.Bind)

	g.server = &http.Server{
		Addr:    g.config.Bind,
		Handler: g.router.GetMux(),
	}

	return g.server.ListenAndServe()
}

// 关闭GoStar
func (g *goStar) Close() error {
	logger.Close()
	return g.server.Close()
}

// 获取GoStar上下文
func GetContext() *goStar {
	return instance
}
