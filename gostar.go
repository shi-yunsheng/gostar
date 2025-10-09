package gostar

import (
	"net/http"
	"sync"

	"github.com/shi-yunsheng/gostar/logger"
	"github.com/shi-yunsheng/gostar/model"
	"github.com/shi-yunsheng/gostar/router"
	"github.com/shi-yunsheng/gostar/router/handler"
	"github.com/shi-yunsheng/gostar/router/middleware"
)

var (
	instance *goStar
	once     sync.Once
)

type goStar struct {
	version string
	config  *config
	server  *http.Server
	router  *router.Router
}

// @en new goStar instance
//
// @zh 新建GoStar实例
func New() *goStar {
	instance := &goStar{
		version: "1.0.0",
		config:  getConfig(),
		router:  router.NewRouter(),
	}
	instance.initGoStar()
	return instance
}

// @en new goStar instance once
//
// @zh 新建唯一GoStar实例
func NewOnce() *goStar {
	once.Do(func() {
		instance = &goStar{
			version: "1.0.0",
			config:  getConfig(),
			router:  router.NewRouter(),
		}
		instance.initGoStar()
	})

	return instance
}

// @en init goStar
//
// @zh 初始化GoStar
func (g *goStar) initGoStar() {
	// @en if debug mode is enabled, enable debug mode in all components
	// @zh 如果调试模式已开启，则开启各个组件的调试模式
	if g.config.Debug {
		middleware.EnableDebug()
		handler.EnableDebug()
	}

	g.initDate()
	g.initLog()
	model.InitDB(g.config.Database)
	model.InitRedis(g.config.Redis)

	// @en use default router middleware
	//
	// @zh 使用默认路由中间件
	g.router.UseMiddleware(
		middleware.ErrorMiddleware,
		middleware.LogMiddleware,
		middleware.CORSMiddleware(g.config.AllowedOrigins),
	)
}

// @en returns the version of the goStar
//
// @zh 返回GoStar的版本
func (g *goStar) Version() string {
	return g.version
}

// @en runs the goStar
//
// @zh 运行GoStar
func (g *goStar) Run() error {
	logger.I("GoStar is running on " + g.config.Bind)

	g.server = &http.Server{
		Addr:    g.config.Bind,
		Handler: g.router.GetMux(),
	}

	return g.server.ListenAndServe()
}

// @en close goStar
//
// @zh 关闭GoStar
func (g *goStar) Close() error {
	logger.Close()
	return g.server.Close()
}
