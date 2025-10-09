package router

import (
	"gostar/date"
	"gostar/router/handler"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// @en parse route
//
// @zh 解析路由
func (r *Router) parseRoute(routes []Route, parent string) {
	for i := range routes {
		route := &routes[i]

		// @zh 如果Webapp和Static同时设置，则抛出错误
		if route.Webapp != nil && route.Static != nil {
			panic("Webapp and Static cannot be set at the same time")
		}

		// @zh 如果handler、webapp、static、websocket和children都为空，则抛出错误
		if route.Handler == nil && route.Webapp == nil && route.Static == nil && !route.Websocket && len(route.Children) == 0 {
			panic("handler, webapp, static, websocket and children cannot all be empty")
		}

		if !strings.HasPrefix(route.Path, "/") {
			route.Path = "/" + route.Path
		}

		// @en exclude "/", otherwise it will conflict with the root path
		// @zh 排除"/"，否则会和根路径冲突
		if parent != "/" {
			route.parent = parent

			// @en remove ^ and $ from parent path
			// @zh 移除父路径的^和$
			if after, ok := strings.CutPrefix(parent, "^"); ok {
				parent = after
			}
			if after, ok := strings.CutSuffix(parent, "$"); ok {
				parent = after
			}

			route.Path = parent + route.Path
		}

		route.Path, route.params = r.parsePath(route.Path)

		// @en if static or webapp is set, add wildcard matching to the path
		// @zh 如果是Static或Webapp，路径后面加上泛匹配
		if route.Static != nil || route.Webapp != nil {
			// @en add ^ to path
			// @zh 如果路径不以^开头，则添加^
			if !strings.HasPrefix(route.Path, "^") {
				route.Path = "^" + route.Path
			}

			// @en remove $ from parent path
			// @zh 移除父路径末尾的$
			if after, ok := strings.CutSuffix(route.Path, "$"); ok {
				route.Path = after
			}

			// @en static needs to check if directory access is allowed, webapp needs to support SPA
			// @zh static需判断是否允许目录访问，webapp需支持SPA
			if after, ok := strings.CutSuffix(route.Path, "/"); ok {
				route.Path = after
			}
			if route.Static != nil && route.Static.AllowDir || route.Webapp != nil {
				route.Path = route.Path + `(/.*)?$`
			} else {
				route.Path = route.Path + `(/[^/]+.*)?$`
			}
		}

		// @en store route
		// @zh 存储路由
		r.routes[route.Path] = route

		// @en if static file or webapp is not set, parse children
		// @zh 如果静态文件或网站设置为空，则解析子路由
		if route.Static == nil && route.Webapp == nil && len(route.Children) > 0 {
			r.parseRoute(route.Children, route.Path)
		}
	}
}

// @en parse path
//
// @zh 解析路径
func (r *Router) parsePath(path string) (string, []handler.Param) {
	// @zh 没有路径参数，直接返回路径
	if !(strings.Contains(path, "{") && strings.Contains(path, "}")) {
		return path, nil
	}

	// @en match path parameters
	// @zh 匹配路径参数
	re := regexp.MustCompile(`\{([^:{}]+)(?::([^:{}]+))?(?::([^:{}]+))?\}`)
	allMatches := re.FindAllStringSubmatch(path, -1)

	// @en if no matches found, return original path
	// @zh 如果没有找到匹配，返回原始路径
	if len(allMatches) == 0 {
		return path, nil
	}

	// @en store path parameters information
	// @zh 存储当前路径的所有参数信息
	params := make([]handler.Param, 0)

	for _, matches := range allMatches {
		// @en remove empty elements from matches
		// @zh 移除matches中的空元素
		cleanMatches := make([]string, 0, len(matches))
		for _, m := range matches {
			if m != "" {
				cleanMatches = append(cleanMatches, m)
			}
		}
		matches = cleanMatches

		// @en the number of matches is not correct
		// @zh 匹配的数量不正确
		if len(matches) > 4 || len(matches) < 2 {
			panic("invalid path parameter format, must be {param[:type:default]}, e.g.: {id}, {page:int:1}, {cid:323}. Got: " + path)
		}

		var (
			param        = matches[1]
			typer        string
			defaultValue any
			optional     bool
		)

		// @en if the parameter name ends with ?, it is considered an optional parameter
		// @zh 如果参数名以?结尾，则认为该参数是可选参数
		if strings.HasSuffix(param, "?") {
			optional = true
			param = strings.TrimSuffix(param, "?")
		}

		switch len(matches) {
		case 2:
			typer = "str"
		case 3:
			switch matches[2] {
			case "int", "float", "str", "bool", "date":
				typer = matches[2]
			default:
				typer = "str"
				defaultValue = matches[2]
			}
		case 4:
			switch matches[2] {
			case "int", "float", "str", "bool", "date":
				typer = matches[2]
				if matches[3] != "" {
					switch typer {
					case "int":
						val, err := strconv.ParseInt(matches[3], 10, 64)
						if err != nil {
							// @zh 路径参数默认值无效，必须为整数
							panic("invalid path parameter default value, must be int. Got: " + matches[3])
						}
						defaultValue = val
					case "float":
						val, err := strconv.ParseFloat(matches[3], 64)
						if err != nil {
							// @zh 路径参数默认值无效，必须为浮点数
							panic("invalid path parameter default value, must be float. Got: " + matches[3])
						}
						defaultValue = val
					case "bool":
						val, err := strconv.ParseBool(matches[3])
						if err != nil {
							// @zh 路径参数默认值无效，必须为布尔值
							panic("invalid path parameter default value, must be bool. Got: " + matches[3])
						}
						defaultValue = val
					case "date":
						val, err := date.ParseTimeString(matches[3])
						if err != nil {
							val, err = date.ParseTimestamp(matches[3])
							if err != nil {
								// @zh 路径参数默认值无效，必须为日期格式（如：2025-09-20 20:18:03、2025-09-20）或时间戳（如：1640995200）
								panic("invalid path parameter default value, must be date (e.g.: 2025-09-20 20:18:03 or 2025-09-20) or timestamp (e.g.: 1640995200). Got: " + matches[3])
							}
						}
						defaultValue = val
					default:
						defaultValue = matches[3]
					}
				}
			default:
				// @zh 路径参数类型无效，必须为int, float, str, bool, date
				panic("invalid path parameter type, must be int, float, str, bool, date. Got: " + matches[2])
			}
		}

		var pattern string
		switch typer {
		case "int":
			pattern = `(\d+)`
		case "float":
			pattern = `(\d+\.\d+)`
		case "bool":
			pattern = `(true|false|True|False|TRUE|FALSE|1|0)`
		case "date":
			pattern = `(\d{4}-(?:1[0-2]|0?[1-9])-(?:3[01]|[12][0-9]|0?[1-9])(?:[T_\s](?:[01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9])?)`
		default:
			pattern = `([^/]+)`
		}

		params = append(params, handler.Param{
			Key:      param,
			Type:     typer,
			Default:  defaultValue,
			Pattern:  pattern,
			Optional: optional,
		})
	}

	// @en replace all parameters
	// @zh 替换所有参数
	resultPath := re.ReplaceAllStringFunc(path, func(match string) string {
		for _, param := range params {
			if strings.Contains(match, param.Key) {
				if param.Optional {
					return param.Pattern + "?"
				}
				return param.Pattern
			}
		}
		return `([^/]+)`
	})

	// @en add ^ and $ to prevent general matching
	// @zh 添加^和$防止泛匹配
	resultPath = "^" + resultPath + "$"

	return resultPath, params
}

// @en sort routes by path length (longest first)
//
// @zh 按路径长度排序路由（最长的在前）
func (r *Router) sortRoutes() {
	for path := range r.routes {
		r.sortedRoutes = append(r.sortedRoutes, path)
	}

	sort.Slice(r.sortedRoutes, func(i, j int) bool {
		return len(r.sortedRoutes[i]) > len(r.sortedRoutes[j])
	})
}
