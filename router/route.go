package router

import (
	"encoding/json"
	"net/http"
	"reflect"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/shi-yunsheng/gostar/router/handler"
	"github.com/shi-yunsheng/gostar/router/middleware"

	"github.com/gorilla/websocket"
)

// HTTP方法类型
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

// 路由配置
type Route struct {
	// 允许的HTTP方法，默认不限制
	Method Method
	// 请求路径，支持正则表达式，路径参数。
	// 正则表达式，例如：/user/\d+
	// 路径参数需要用{}包裹，例如：/user/{id}，支持多个路径参数和路径嵌套，例如：/user/{id}/detail/{girlfriend}
	// 路径参数可以指定默认值或类型，格式为：{param:type:default}或{param:default}，例如：/user/{id:int}和/user/list/{page:int:1}，支持类型：int, float, str, bool, date, default: str
	// 路径参数支持可选，参数名使用?结尾表示可选，例如：/user/list/{page?:int}
	Path string
	// 认证密钥，如果设置，则请求头中必须包含该密钥，否则会返回401错误，例如：{"secret": "aha~"}
	SecretKey map[string]string
	// 请求处理函数
	Handler handler.Handler
	// 子路由
	Children []Route
	// 中间件函数，洋葱模型
	Middleware []middleware.Middleware
	// 是否是WebSocket连接
	Websocket bool
	// websocket升级配置，只有在Websocket为true时有效，不设置时，使用默认配置
	WebsocketUpgrade *websocket.Upgrader
	// 静态文件配置，该选项设置后，Children将无效，Handler则用于前置处理，可中断后续流程
	// 注：Webapp和Static不能同时设置
	Static *handler.Static
	// 网站配置，默认支持SPA，该选项设置后，Children将无效，Handler则用于前置处理，可中断后续流程
	// 注：Webapp和Static不能同时设置。
	// 注：在SPA下，对应路径下的未知路由将交给webapp处理
	Webapp *handler.Webapp
	// 路径参数
	params []handler.Param
	// 父路径
	parent string
	// 模型，可以实现"Validate() error"接口，如果有"Validate"接口，则优先使用"Validate"接口进行校验，
	// 否则使用 github.com/go-playground/validator/v10 进行校验，有关validator的用法请参考 https://github.com/go-playground/validator
	Bind any
	// 确保模型类型只初始化一次
	once sync.Once
	// 缓存模型类型
	modelType reflect.Type
}

// 验证绑定参数
func (r *Route) Validate(req *handler.Request) (any, error) {
	// 初始化模型类型
	r.once.Do(func() {
		t := reflect.TypeOf(r.Bind)
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		r.modelType = t
	})
	// 创建模型实例
	modelInstance := reflect.New(r.modelType).Interface()
	// 如果没有指定请求方式，则根据实际的请求方式进行绑定，POST、PUT、PATCH、DELETE请求方式绑定请求体，其他请求方式绑定查询参数
	if r.Method == "" {
		var model map[string]any

		switch r.Method {
		case "POST", "PUT", "PATCH", "DELETE":
			body, err := req.GetAllBody()
			if err != nil {
				return nil, err
			}
			model = body
		default:
			query := req.GetAllQuery()
			model = query
		}

		jsonData, err := json.Marshal(model)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(jsonData, modelInstance)
		if err != nil {
			return nil, err
		}
		// 如果模型实现了"Validate() error"接口，则使用Validate接口进行校验
		if validateObj, ok := modelInstance.(interface{ Validate() error }); ok {
			err = validateObj.Validate()
			if err != nil {
				return nil, err
			}
		} else {
			// 模型没有实现"Validate() error"接口，使用github.com/go-playground/validator/v10进行校验
			validate := validator.New()
			err = validate.Struct(modelInstance)
			if err != nil {
				return nil, err
			}
		}
	}
	return modelInstance, nil
}

// 路由器管理HTTP路由和处理器
type Router struct {
	// HTTP ServeMux实例
	mux *http.ServeMux
	// 路由表
	routes map[string]*Route
	// 排序后的路由
	sortedRoutes []string
	// 全局中间件，洋葱模型
	middleware []middleware.Middleware
	// 全局认证密钥，如果设置，则请求头中必须包含该密钥，否则会返回401错误，例如：{"secret": "aha~"}
	// 如果和路由的SecretKey都包含相同Key，则优先使用路由的SecretKey
	secretKey map[string]string
}

// 获取HTTP ServeMux实例，使用它来设置HTTP服务器
func (r *Router) GetMux() *http.ServeMux {
	return r.mux
}

// 获取路由表
func (r *Router) GetRoutes() map[string]*Route {
	return r.routes
}
