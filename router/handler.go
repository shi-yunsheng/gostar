package router

import (
	"regexp"
	"strings"

	"github.com/shi-yunsheng/gostar/router/handler"
)

// @en get method from route recursively (child has priority, if empty, use parent)
// @zh 递归获取路由的Method（子路由优先，如果为空则使用父路由）
func (r *Router) getMethod(route *Route) Method {
	if route.Method != "" {
		return route.Method
	}

	if route.parent != "" && r.routes[route.parent] != nil {
		return r.getMethod(r.routes[route.parent])
	}

	return ""
}

// @en parse param
//
// @zh 解析参数
func (r *Router) parseParam(route *Route, path string) []handler.Param {
	// @en merge parent route and current route params
	// @zh 合并父路由和当前路由的参数
	params := make([]handler.Param, 0)
	if route.parent != "" {
		params = append(params, r.routes[route.parent].params...)
	}
	params = append(params, route.params...)

	// @en regex match params
	// @zh 正则匹配参数
	re := regexp.MustCompile(route.Path)
	allMatches := re.FindAllStringSubmatch(path, -1)

	if len(allMatches) == 0 {
		return nil
	}

	matches := allMatches[0][1:]

	// @en set default value
	// @zh 预设默认值
	if len(matches) < len(params) {
		for i := len(matches); i < len(params); i++ {
			params[i].Value = params[i].Default
		}
	}

	// @en set parameter value
	// @zh 设置参数值
	for i, match := range matches {
		// @zh 如果参数数量小于匹配数量
		if i >= len(params) {
			// @zh 如果是static或webapp，将后面的视为文件路径
			if route.Static != nil || route.Webapp != nil {
				params = append(params, handler.Param{
					Key:   "__filepath__",
					Value: match,
				})
			}

			break
		}

		reg := regexp.MustCompile(params[i].Pattern)
		if reg.MatchString(match) {
			params[i].Value = match
		} else if !params[i].Optional {
			panic("invalid path parameter value, must be " + params[i].Type + ". Got: " + match)
		}
	}

	return params
}

// @en root handler, all requests will pass through here
//
// @zh 根处理器，所有请求都会经过这里
func (r *Router) serveHTTP(w *handler.Response, req handler.Request) {
	path := req.URL.Path

	route, ok := r.routes[path]

	// @en if not found, match by regex
	// @zh 如果获取不到，进行正则匹配
	if !ok {
		// @en remove / from path
		// @zh 如果路径以/结尾，则去掉/
		if after, ok := strings.CutSuffix(path, "/"); ok {
			path = after
		}

		for _, rt := range r.sortedRoutes {
			re := regexp.MustCompile(r.routes[rt].Path)
			if re.MatchString(path) {
				route = r.routes[rt]
				break
			}
		}
	}

	if route == nil {
		handler.NotFound(w, req)
		return
	}

	// @en validate request method (recursively check parent routes)
	// @zh 验证请求方式（递归检查父路由）
	method := r.getMethod(route)
	if method != "" && string(method) != req.Method {
		handler.MethodNotAllowed(w, req)
		return
	}

	// @en validate SecretKey
	// @zh 验证SecretKey
	if route.SecretKey != nil {
		for key, value := range route.SecretKey {
			if req.GetHeader(key) != value {
				handler.Unauthorized(w, req)
				return
			}
		}
	}

	req.SetParams(r.parseParam(route, path))
	if route.Bind != nil {
		model, err := route.Validate(&req)
		if err != nil {
			handler.BadRequest(w, req, err)
			return
		}
		req.SetBindModel(model)
	}

	handlerFunc := route.Handler

	// @en use path middleware
	// @zh 使用路径中间件
	for i := len(route.Middleware) - 1; i >= 0; i-- {
		handlerFunc = route.Middleware[i](handlerFunc)
	}

	if route.Webapp != nil {
		handlerFunc = handler.WebApp(handlerFunc, route.Webapp)
	} else if route.Static != nil {
		handlerFunc = handler.StaticServer(handlerFunc, route.Static)
	} else if route.Websocket {
		handlerFunc = handler.ToWebsocketHandler(handlerFunc, route.WebsocketUpgrade)
	}

	// @en if handlerFunc is nil, return 404
	// @zh 如果handlerFunc为nil，返回404
	if handlerFunc == nil {
		handler.NotFound(w, req)
		return
	}

	handlerFunc(w, req)
}
