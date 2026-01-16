package router

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/shi-yunsheng/gostar/date"
	"github.com/shi-yunsheng/gostar/router/handler"
	"github.com/shi-yunsheng/gostar/router/middleware"
)

// 解析路由
func (r *Router) parseRoute(routes []Route, parentPath, parentRegex string) {
	for i := range routes {
		route := &routes[i]
		// 如果Webapp和Static同时设置，则抛出错误
		if route.Webapp != nil && route.Static != nil {
			panic("Webapp and Static cannot be set at the same time")
		}
		// 如果handler、webapp、static、websocket和children都为空，则抛出错误
		if route.Handler == nil && route.Webapp == nil && route.Static == nil && !route.Websocket && len(route.Children) == 0 {
			panic("handler, webapp, static, websocket and children cannot all be empty")
		}

		// 构建完整路径
		var fullPath string
		if parentPath != "" && parentPath != "/" {
			fullPath = strings.TrimSuffix(parentPath, "/") + "/" + strings.TrimPrefix(route.Path, "/")
		} else {
			fullPath = route.Path
		}

		// 确保路径以/开头
		if !strings.HasPrefix(fullPath, "/") {
			fullPath = "/" + fullPath
		}

		// 保存原始路径
		route.originalPath = fullPath

		// 解析路径为正则表达式
		regexPath, params := r.parsePath(fullPath)
		route.params = params

		// 如果是Static或Webapp，路径后面加上泛匹配
		if route.Static != nil || route.Webapp != nil {
			// 确保路径以^开头
			if !strings.HasPrefix(regexPath, "^") {
				regexPath = "^" + regexPath
			}
			// 移除末尾的$
			if after, ok := strings.CutSuffix(regexPath, "$"); ok {
				regexPath = after
			}
			// 移除末尾的/
			if after, ok := strings.CutSuffix(regexPath, "/"); ok {
				regexPath = after
			}
			if route.Static != nil && route.Static.AllowDir || route.Webapp != nil {
				regexPath = regexPath + `(/.*)?$`
			} else {
				regexPath = regexPath + `(/[^/]+.*)?$`
			}
		}

		// 存储路由，使用正则表达式作为key
		r.routes[regexPath] = route
		route.regexPath = regexPath

		// 合并父路由的配置（SecretKey和Middleware）
		if parentRegex != "" && r.routes[parentRegex] != nil {
			parentRoute := r.routes[parentRegex]
			// 从父路由合并SecretKey到子路由
			if parentRoute.SecretKey != nil {
				if route.SecretKey == nil {
					route.SecretKey = make(map[string]string)
				}
				for key, value := range parentRoute.SecretKey {
					// 如果子路由有相同的key，保留子路由的值
					if _, exists := route.SecretKey[key]; !exists {
						route.SecretKey[key] = value
					}
				}
			}
			// 从父路由合并中间件到子路由（父路由中间件在前）
			if len(parentRoute.Middleware) > 0 {
				mergedMiddleware := make([]middleware.Middleware, 0, len(parentRoute.Middleware)+len(route.Middleware))
				mergedMiddleware = append(mergedMiddleware, parentRoute.Middleware...)
				mergedMiddleware = append(mergedMiddleware, route.Middleware...)
				route.Middleware = mergedMiddleware
			}
		}

		// 如果静态文件或网站设置为空，则解析子路由
		if route.Static == nil && route.Webapp == nil && len(route.Children) > 0 {
			r.parseRoute(route.Children, fullPath, regexPath)
		}
	}
}

// 解析路径
func (r *Router) parsePath(path string) (string, []handler.Param) {
	// 如果没有路径参数，返回转义后的正则路径
	if !strings.Contains(path, "{") || !strings.Contains(path, "}") {
		return "^" + regexp.QuoteMeta(path) + "$", nil
	}

	// 匹配路径参数
	re := regexp.MustCompile(`\{([^:{}]+)(?::([^:{}]+))?(?::([^:{}]+))?\}`)
	allMatches := re.FindAllStringSubmatch(path, -1)
	// 如果没有找到匹配，返回原始路径
	if len(allMatches) == 0 {
		return "^" + regexp.QuoteMeta(path) + "$", nil
	}
	// 存储当前路径的所有参数信息
	params := make([]handler.Param, 0)

	for _, matches := range allMatches {
		// 移除matches中的空元素
		cleanMatches := make([]string, 0, len(matches))
		for _, m := range matches {
			if m != "" {
				cleanMatches = append(cleanMatches, m)
			}
		}
		matches = cleanMatches
		// 匹配的数量不正确
		if len(matches) > 4 || len(matches) < 2 {
			panic("invalid path parameter format, must be {param[:type:default]}, e.g.: {id}, {page:int:1}, {cid:323}. Got: " + path)
		}

		var (
			param        = matches[1]
			typer        string
			defaultValue any
			optional     bool
		)
		// 如果参数名以?结尾，则认为该参数是可选参数
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
							// 路径参数默认值无效，必须为整数
							panic("invalid path parameter default value, must be int. Got: " + matches[3])
						}
						defaultValue = val
					case "float":
						val, err := strconv.ParseFloat(matches[3], 64)
						if err != nil {
							// 路径参数默认值无效，必须为浮点数
							panic("invalid path parameter default value, must be float. Got: " + matches[3])
						}
						defaultValue = val
					case "bool":
						val, err := strconv.ParseBool(matches[3])
						if err != nil {
							// 路径参数默认值无效，必须为布尔值
							panic("invalid path parameter default value, must be bool. Got: " + matches[3])
						}
						defaultValue = val
					case "date":
						val, err := date.ParseTimeString(matches[3])
						if err != nil {
							val, err = date.ParseTimestamp(matches[3])
							if err != nil {
								// 路径参数默认值无效，必须为日期格式（如：2025-09-20 20:18:03、2025-09-20）或时间戳（如：1640995200）
								panic("invalid path parameter default value, must be date (e.g.: 2025-09-20 20:18:03 or 2025-09-20) or timestamp (e.g.: 1640995200). Got: " + matches[3])
							}
						}
						defaultValue = val
					default:
						defaultValue = matches[3]
					}
				}
			default:
				// 路径参数类型无效，必须为int, float, str, bool, date
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
	// 使用正则表达式替换所有参数
	resultBuilder := strings.Builder{}
	lastIndex := 0
	matchIndices := re.FindAllStringSubmatchIndex(path, -1)

	for idx, matchIdx := range matchIndices {
		start := matchIdx[0]
		end := matchIdx[1]
		param := params[idx]

		literal := path[lastIndex:start]
		// 转义字面量部分中的正则元字符
		literal = regexp.QuoteMeta(literal)

		if param.Optional && strings.HasSuffix(literal, `\/`) {
			// 如果可选参数前有"/"，则将"/"和参数一起设为可选
			literal = strings.TrimSuffix(literal, `\/`)
			resultBuilder.WriteString(literal)

			innerPattern := strings.TrimSuffix(strings.TrimPrefix(param.Pattern, "("), ")")
			if innerPattern != param.Pattern {
				resultBuilder.WriteString("(?:/(")
				resultBuilder.WriteString(innerPattern)
				resultBuilder.WriteString("))?")
			} else {
				// 理论上不会发生，因为所有模式都包含括号
				resultBuilder.WriteString(param.Pattern + "?")
			}
		} else {
			resultBuilder.WriteString(literal)
			if param.Optional {
				resultBuilder.WriteString(param.Pattern + "?")
			} else {
				resultBuilder.WriteString(param.Pattern)
			}
		}

		lastIndex = end
	}

	// 添加剩余部分并转义
	remaining := path[lastIndex:]
	if remaining != "" {
		resultBuilder.WriteString(regexp.QuoteMeta(remaining))
	}

	resultPath := resultBuilder.String()
	// 添加^和$防止泛匹配
	resultPath = "^" + resultPath + "$"

	return resultPath, params
}

// 按路径类型和长度排序路由
func (r *Router) sortRoutes() {
	r.sortedRoutes = make([]string, 0, len(r.routes))
	for path := range r.routes {
		r.sortedRoutes = append(r.sortedRoutes, path)
	}

	// 定义正则元字符集合
	regexpMetaChars := map[byte]bool{
		'^': true, '$': true, '.': true, '*': true, '+': true, '?': true,
		'|': true, '(': true, ')': true, '[': true, ']': true, '{': true,
		'}': true, '\\': true,
	}

	sort.Slice(r.sortedRoutes, func(i, j int) bool {
		a, b := r.sortedRoutes[i], r.sortedRoutes[j]

		// 检查是否为精确路径（不包含正则元字符，除了^和$）
		isExactPathA := true
		isExactPathB := true

		for k := 0; k < len(a); k++ {
			if a[k] == '^' && k == 0 {
				continue // 开头的^不算正则
			}
			if a[k] == '$' && k == len(a)-1 {
				continue // 结尾的$不算正则
			}
			if regexpMetaChars[a[k]] {
				isExactPathA = false
				break
			}
		}

		for k := 0; k < len(b); k++ {
			if b[k] == '^' && k == 0 {
				continue // 开头的^不算正则
			}
			if b[k] == '$' && k == len(b)-1 {
				continue // 结尾的$不算正则
			}
			if regexpMetaChars[b[k]] {
				isExactPathB = false
				break
			}
		}

		// 精确路径优先于正则路径
		if isExactPathA && !isExactPathB {
			return true
		}
		if !isExactPathA && isExactPathB {
			return false
		}

		// 同类型中，按长度降序排列
		if len(a) != len(b) {
			return len(a) > len(b)
		}

		// 长度相同，按字母顺序稳定排序
		return a < b
	})
}
