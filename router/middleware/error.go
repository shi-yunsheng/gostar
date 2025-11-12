package middleware

import (
	"fmt"

	"github.com/shi-yunsheng/gostar/logger"
	"github.com/shi-yunsheng/gostar/router/handler"
	"github.com/shi-yunsheng/gostar/utils"
)

var debug = false

// 开启调试模式
func EnableDebug() {
	debug = true
}

// 错误处理中间件
func ErrorMiddleware(next handler.Handler) handler.Handler {
	return func(w *handler.Response, r handler.Request) any {
		defer func() {
			if err := recover(); err != nil {
				logger.E("Error: %v", err)
				// 如果在调试模式下，也将堆栈信息返回给客户端
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
						if r.Method == "GET" {
							handler.InternalServerError(w, r, fmt.Errorf("internal server error: %v", err))
						} else {
							w.Json(map[string]any{
								"error":   "internal server error",
								"message": err,
								"stack":   stackInfo,
							})
						}
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
					if r.Method == "GET" {
						handler.InternalServerError(w, r, fmt.Errorf("internal server error: %v", err))
					} else {
						w.Json(map[string]any{
							"error":   "internal server error",
							"message": err,
						})
					}
				}
			}
		}()
		// 继续处理请求并返回结果
		return next(w, r)
	}
}
