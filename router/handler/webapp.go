package handler

import (
	"net/http"
	"strings"

	"github.com/shi-yunsheng/gostar/utils"
)

// webapp配置
type Webapp struct {
	// 首页文件
	Index string
	// 网站根路径
	Path string
	// 资源路径，如果设置，则优先使用该路径，否则使用“./web”
	AssetsPath string
	// 是否禁用SPA，如果设置为true，单页应用路由将失效
	DisableSPA bool
}

var (
	webappIndex string = "index.html"
	webappPath  string = "./web"
	assetsPath  string = ""
	hasCache    bool   = false
)

// webapp处理器
func WebApp(handler Handler, webconfig *Webapp) Handler {
	// 如果webconfig不为nil，则设置index和path
	if webconfig != nil && !hasCache {
		if webconfig.Index != "" {
			webappIndex = webconfig.Index
		}
		if webconfig.Path != "" {
			webappPath = webconfig.Path
		}
		if webconfig.AssetsPath != "" {
			assetsPath = webconfig.AssetsPath
		}
		hasCache = true
	}

	return func(w *Response, r Request) any {
		if handler != nil {
			handler(w, r)
		}
		// 如果在handler中调用了EarlyBreak，则不进行静态文件服务
		if w.GetEarlyBreak() {
			return nil
		}

		filePath := r.GetParam("__filepath__")
		if filepath, ok := filePath.(string); ok && filepath != "" {
			path := webappPath
			if assetsPath != "" {
				path = assetsPath
			}
			// 如果文件存在，则服务它
			if utils.IsExists(path+filepath) && utils.IsFile(path+filepath) {
				http.ServeFile(w, r.Request, path+filepath)
				return nil
			} else {
				// 如果SPA被禁用，则返回404
				if webconfig.DisableSPA {
					NotFound(w, r)
					return nil
				}
			}
		}

		if !strings.HasPrefix(webappIndex, "/") {
			webappIndex = "/" + webappIndex
		}

		http.ServeFile(w, r.Request, webappPath+webappIndex)
		return nil
	}
}
