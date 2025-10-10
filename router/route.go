package router

import (
	"net/http"

	"github.com/shi-yunsheng/gostar/router/handler"
	"github.com/shi-yunsheng/gostar/router/middleware"

	"github.com/gorilla/websocket"
)

// @en HTTP method type
//
// @zh HTTP方法类型
type Method string

const (
	GET     Method = "GET"
	POST    Method = "POST"
	PUT     Method = "PUT"
	DELETE  Method = "DELETE"
	PATCH   Method = "PATCH"
	OPTIONS Method = "OPTIONS"
	HEAD    Method = "HEAD"
)

// @en Route configuration
//
// @zh 路由配置
type Route struct {
	// @en HTTP method, default no limit
	//
	// @zh 允许的HTTP方法，默认不限制
	Method Method
	// @en request path, supports regular expressions, path parameters, and nested paths
	// Regular expressions, e.g.: /user/\d+
	// Path parameters must be wrapped in {}, e.g.: /user/{id}, supports multiple path parameters and nested paths, e.g.: /user/{id}/detail/{girlfriend}
	// Path parameters can specify default values or types, format: {param:type:default} or {param:default}, e.g.: /user/{id:int} and /user/list/{page:int:1}, supported types: int, float, str, bool, date, default: str
	// Path parameters support optional, the parameter name uses ? to indicate optional, e.g.: /user/list/{page?:int}
	//
	// @zh 请求路径，支持正则表达式，路径参数。
	// 正则表达式，例如：/user/\d+
	// 路径参数需要用{}包裹，例如：/user/{id}，支持多个路径参数和路径嵌套，例如：/user/{id}/detail/{girlfriend}
	// 路径参数可以指定默认值或类型，格式为：{param:type:default}或{param:default}，例如：/user/{id:int}和/user/list/{page:int:1}，支持类型：int, float, str, bool, date, default: str
	// 路径参数支持可选，参数名使用?结尾表示可选，例如：/user/list/{page?:int}
	Path string
	// @en secret key for authentication, if set, the request header must contain the key, otherwise it will return a 401 error, for example: {"secret": "aha~"}
	//
	// @zh 认证密钥，如果设置，则请求头中必须包含该密钥，否则会返回401错误，例如：{"secret": "aha~"}
	SecretKey map[string]string
	// @en request handler function
	//
	// @zh 请求处理函数
	Handler handler.Handler
	// @en children routes
	//
	// @zh 子路由
	Children []Route
	// @en middleware functions, onion model
	//
	// @zh 中间件函数，洋葱模型
	Middleware []middleware.Middleware
	// @en whether it's a WebSocket connection
	//
	// @zh 是否是WebSocket连接
	Websocket bool
	// @en websocket upgrade, only valid when Websocket is true, if not set, use default configuration
	//
	// @zh websocket升级配置，只有在Websocket为true时有效，不设置时，使用默认配置
	WebsocketUpgrade *websocket.Upgrader
	// @en static file configuration, if set, Children will be invalid, Handler will be used for pre-processing, and the subsequent process can be interrupted
	// tip: Webapp and Static cannot be set at the same time
	//
	// @zh 静态文件配置，该选项设置后，Children将无效，Handler则用于前置处理，可中断后续流程
	// 注：Webapp和Static不能同时设置
	Static *handler.Static
	// @en webapp configuration, default supports SPA, if set, Children will be invalid, Handler will be used for pre-processing, and the subsequent process can be interrupted
	// tip: Webapp and Static cannot be set at the same time
	// tip: in SPA, unknown routes under the corresponding path will be handled by webapp
	//
	// @zh 网站配置，默认支持SPA，该选项设置后，Children将无效，Handler则用于前置处理，可中断后续流程
	// 注：Webapp和Static不能同时设置。
	// 注：在SPA下，对应路径下的未知路由将交给webapp处理
	Webapp *handler.Webapp
	// @en path parameters
	//
	// @zh 路径参数
	params []handler.Param
	// @en parent path
	//
	// @zh 父路径
	parent string
	// @en parameter binding, used for deserializing parameters
	//
	// @zh 参数绑定，用于反序列化参数
	Bind *handler.Bind
}

// @en Router manages HTTP routes and handlers
//
// @zh 路由器管理HTTP路由和处理器
type Router struct {
	// @en HTTP ServeMux instance
	//
	// @zh HTTP ServeMux实例
	mux *http.ServeMux
	// @en routes table
	//
	// @zh 路由表
	routes map[string]*Route
	// @en sorted routes
	//
	// @zh 排序后的路由
	sortedRoutes []string
	// @en global middleware, onion model
	//
	// @zh 全局中间件，洋葱模型
	middleware []middleware.Middleware
	// @en global secret key for authentication, if set, the request header must contain the key, otherwise it will return a 401 error, for example: {"secret": "aha~"}
	// If the Key is the same as the SecretKey in the route, the route's SecretKey will be used
	//
	// @zh 全局认证密钥，如果设置，则请求头中必须包含该密钥，否则会返回401错误，例如：{"secret": "aha~"}
	// 如果和路由的SecretKey都包含相同Key，则优先使用路由的SecretKey
	secretKey map[string]string
}

// @en get HTTP ServeMux instance, use it to set up the HTTP server
//
// @zh 获取HTTP ServeMux实例，使用它来设置HTTP服务器
func (r *Router) GetMux() *http.ServeMux {
	return r.mux
}

// @en get routes table
//
// @zh 获取路由表
func (r *Router) GetRoutes() map[string]*Route {
	return r.routes
}
