package router

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/shi-yunsheng/gostar/date"
	"github.com/shi-yunsheng/gostar/router/handler"
)

// 递归获取路由的Method（子路由优先，如果为空则使用父路由）
func (r *Router) getMethod(route *Route) Method {
	if route.Method != "" {
		return route.Method
	}

	if route.parent != "" && r.routes[route.parent] != nil {
		return r.getMethod(r.routes[route.parent])
	}

	return ""
}

// 解析参数
func (r *Router) parseParam(route *Route, path string) []handler.Param {
	// 合并父路由和当前路由的参数
	params := make([]handler.Param, 0)
	if route.parent != "" {
		params = append(params, r.routes[route.parent].params...)
	}
	params = append(params, route.params...)
	// 正则匹配参数
	re := regexp.MustCompile(route.Path)
	allMatches := re.FindAllStringSubmatch(path, -1)

	if len(allMatches) == 0 {
		return nil
	}

	matches := allMatches[0][1:]
	// 预设默认值
	if len(matches) < len(params) {
		for i := len(matches); i < len(params); i++ {
			params[i].Value = params[i].Default
		}
	}
	// 设置参数值
	for i, match := range matches {
		// 如果参数数量小于匹配数量
		if i >= len(params) {
			// 如果是static或webapp，将后面的视为文件路径
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
			// 根据类型进行转换
			convertedValue, err := convertParamValue(match, params[i].Type)
			if err != nil {
				if !params[i].Optional {
					panic("invalid path parameter value, must be " + params[i].Type + ". Got: " + match + ". Error: " + err.Error())
				}
			} else {
				params[i].Value = convertedValue
			}
		} else if !params[i].Optional {
			panic("invalid path parameter value, must be " + params[i].Type + ". Got: " + match)
		}
	}

	return params
}

// 转换参数值
func convertParamValue(value string, paramType string) (any, error) {
	switch paramType {
	case "int":
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return val, nil
	case "float":
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return val, nil
	case "bool":
		val, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return val, nil
	case "date":
		val, err := date.ParseTimeString(value)
		if err != nil {
			val, err = date.ParseTimestamp(value)
			if err != nil {
				return nil, err
			}
		}
		return val, nil
	case "str", "":
		// 默认类型为字符串
		return value, nil
	default:
		// 未知类型，返回原值
		return value, nil
	}
}

// 根处理器，所有请求都会经过这里
func (r *Router) serveHTTP(w *handler.Response, req handler.Request) any {
	path := req.URL.Path

	route, ok := r.routes[path]
	// 如果获取不到，进行正则匹配
	if !ok {
		// 如果路径以/结尾，则去掉/
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
		return nil
	}

	// 验证请求方式（递归检查父路由）
	method := r.getMethod(route)
	if method != "" && string(method) != req.Method {
		handler.MethodNotAllowed(w, req)
		return nil
	}
	// 验证SecretKey
	if r.secretKey != nil {
		for key, value := range r.secretKey {
			if req.GetHeader(key) != value {
				handler.Unauthorized(w, req)
				return nil
			}
		}
	}
	if route.SecretKey != nil {
		for key, value := range route.SecretKey {
			if req.GetHeader(key) != value {
				handler.Unauthorized(w, req)
				return nil
			}
		}
	}

	req.SetParams(r.parseParam(route, path))
	if route.Bind != nil {
		model, err := route.Validate(&req)
		if err != nil {
			if model != nil {
				return model
			}

			handler.BadRequest(w, req, err)
			return nil
		}
		req.SetBindModel(model)
	}

	handlerFunc := route.Handler
	// 使用路径中间件
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
	// 如果handlerFunc为nil，返回404
	if handlerFunc == nil {
		handler.NotFound(w, req)
		return nil
	}
	// 调用handler并返回结果
	return handlerFunc(w, req)
}
