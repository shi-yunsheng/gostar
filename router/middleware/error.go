package middleware

import (
	"fmt"

	"github.com/shi-yunsheng/gostar/logger"
	"github.com/shi-yunsheng/gostar/router/handler"
	"github.com/shi-yunsheng/gostar/utils"
)

var debug = false

// @en enable debug mode
//
// @zh 开启调试模式
func EnableDebug() {
	debug = true
}

// @en error middleware
//
// @zh 错误处理中间件
func ErrorMiddleware(next handler.Handler) handler.Handler {
	return func(w *handler.Response, r handler.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.E("Error: %v", err)

				// @en if in debug mode, also return stack info to client
				// @zh 如果在调试模式下，也将堆栈信息返回给客户端
				if debug {
					stackInfo := utils.GetStackTrace()

					logger.E("Stack Info:\n%s", stackInfo)

					if r.IsWebsocket() {
						conn := w.GetWebsocketConn()
						conn.SendJson(handler.ResponseBody{
							Code:    0,
							Message: fmt.Sprintf("internal server error: %v\n\nstack info:\n%s", err, stackInfo),
							Data:    nil,
						})
					} else {
						w.Error(fmt.Errorf("internal server error: %v\n\nstack info: %s", err, stackInfo))
					}
					return
				}

				if r.IsWebsocket() {
					conn := w.GetWebsocketConn()
					conn.SendJson(handler.ResponseBody{
						Code:    0,
						Message: fmt.Sprintf("internal server error: %v", err),
						Data:    nil,
					})
				} else {
					w.Error(fmt.Errorf("internal server error: %v", err))
				}
			}
		}()

		// 继续处理请求
		next(w, r)
	}
}
