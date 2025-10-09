package handler

import (
	"net/http"
	"strings"

	"github.com/shi-yunsheng/gostar/utils"
)

// @en webapp configuration
//
// @zh webapp配置
type Webapp struct {
	// @en index page file
	//
	// @zh 首页文件
	Index string
	// @en webapp root path
	//
	// @zh 网站根路径
	Path string
	// @en assets path, if set, will be used instead of "./web"
	//
	// @zh 资源路径，如果设置，则优先使用该路径，否则使用“./web”
	AssetsPath string
	// @en whether to disable SPA; if set to true, SPA routing will be invalid
	//
	// @zh 是否禁用SPA，如果设置为true，单页应用路由将失效
	DisableSPA bool
}

var (
	webappIndex string = "index.html"
	webappPath  string = "./web"
	assetsPath  string = ""
	hasCache    bool   = false
)

// @en webapp handler
//
// @zh webapp处理器
func WebApp(handler Handler, webconfig *Webapp) Handler {
	// @en if webconfig is not nil, set index and path
	// @zh 如果webconfig不为nil，则设置index和path
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

	return func(w *Response, r Request) {
		if handler != nil {
			handler(w, r)
		}

		// @en if EarlyBreak is called in handler, do not serve static file
		// @zh 如果在handler中调用了EarlyBreak，则不进行静态文件服务
		if w.GetEarlyBreak() {
			return
		}

		filePath := r.GetParam("__filepath__")
		if filepath, ok := filePath.(string); ok && filepath != "" {
			path := webappPath
			if assetsPath != "" {
				path = assetsPath
			}

			// @en if file exists, serve it
			// @zh 如果文件存在，则服务它
			if utils.IsExists(path+filepath) && utils.IsFile(path+filepath) {
				http.ServeFile(w, r.Request, path+filepath)
				return
			} else {
				// @en if SPA is disabled, return 404
				// @zh 如果SPA被禁用，则返回404
				if webconfig.DisableSPA {
					NotFound(w, r)
					return
				}
			}
		}

		if !strings.HasPrefix(webappIndex, "/") {
			webappIndex = "/" + webappIndex
		}

		http.ServeFile(w, r.Request, webappPath+webappIndex)
	}
}
